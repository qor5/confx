package confx

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

var reSplitParams = regexp.MustCompile(`'[^']*'|\S+`)

// parseOneOfParam2 parses a string that contains multiple values separated by
// spaces and/or single quotes. The single quotes are used to enclose values
// that contain spaces.
//
// Examples:
//   - "a b c" -> ["a", "b", "c"]
//   - "'a b' c" -> ["a b", "c"]
//   - "'a b' 'c d'" -> ["a b", "c d"]
//   - "'a b c'" -> ["a b c"]
func parseOneOfParam2(s string) []string {
	vals := reSplitParams.FindAllString(s, -1)
	for i := 0; i < len(vals); i++ {
		vals[i] = strings.ReplaceAll(vals[i], "'", "")
	}
	return vals
}

func requireCheckFieldValue(
	fl validator.FieldLevel, param string, value string, defaultNotFoundValue bool,
) bool {
	field, kind, _, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), param)
	if !found {
		return defaultNotFoundValue
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == cast.ToInt64(value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == cast.ToUint64(value)

	case reflect.Float32:
		return field.Float() == float64(cast.ToFloat32(value))

	case reflect.Float64:
		return field.Float() == cast.ToFloat64(value)

	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) == cast.ToInt64(value)

	case reflect.Bool:
		return field.Bool() == (value == "true")

	case reflect.Ptr:
		if field.IsNil() {
			return value == "nil"
		}
		// Handle non-nil pointers
		return requireCheckFieldValue(fl, param, value, defaultNotFoundValue)
	default:
		// default reflect.String:
		return field.String() == value
	}
}

type Validator interface {
	RegisterValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) error
	StructCtx(ctx context.Context, v any) error
}

type ValidatorFunc func(ctx context.Context, v any) error

type wrappedValidator struct {
	Validator
	structCtxFunc ValidatorFunc
}

func (w *wrappedValidator) StructCtx(ctx context.Context, v any) error {
	return w.structCtxFunc(ctx, v)
}

const (
	// stopIfTag / stopUnlessTag stop validation of a field at that point.
	stopIfTag     = "stop_if"
	stopUnlessTag = "stop_unless"

	// skipNestedUnlessTag is the original name of stop_unless, kept as an alias
	// so existing struct tags keep working.
	//
	// Deprecated: use stop_unless. The "nested" is misleading — the tag is not
	// specific to nested structs (see the note on stopUnlessImpl).
	skipNestedUnlessTag = "skip_nested_unless"
)

// stopTagImpls are registered together by ValidatorWithSkipNestedUnless.
var stopTagImpls = map[string]validator.FuncCtx{
	stopIfTag:           stopIfImpl,
	stopUnlessTag:       stopUnlessImpl,
	skipNestedUnlessTag: stopUnlessImpl,
}

// stopTags are the tags whose "failure" means "stop validating here", not
// "this field is invalid". Their errors are filtered out after StructCtx.
var stopTags = []string{stopIfTag, stopUnlessTag, skipNestedUnlessTag}

// stopUnlessImpl stops validating a field unless every (field, value) pair
// matches. It backs the "stop_unless" tag and its "skip_nested_unless" alias.
//
// stopIfImpl is the same thing with the opposite polarity. Polarity is the ONLY
// difference between the two; see the note there for what "stop" covers.
//
//	type Config struct {
//	  Type   string    `validate:"oneof=local remote"`
//	  Local  LocalConf `validate:"stop_unless=Type local"`
//	  Remote RemoteConf `validate:"stop_unless=Type remote"`
//	}
//
// Local is validated only when Type == "local", Remote only when Type ==
// "remote". All pairs must match for validation to proceed.
//
// Panics if the number of parameters is not even.
func stopUnlessImpl(_ context.Context, fl validator.FieldLevel) bool {
	params := parseOneOfParam2(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for %s %s", fl.GetTag(), fl.FieldName()))
	}
	for i := 0; i < len(params); i += 2 {
		// Returning false is how validation is stopped: it produces an error
		// that the wrapper then filters out by tag name (see stopTags).
		if !requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return false
		}
	}
	return true
}

// stopIfImpl stops validating a field when ANY (field, value) pair matches. It
// backs the "stop_if" tag.
//
// "stop" rather than "skip", because what it stops depends on where the tag
// sits, and both are the same underlying behaviour — validator abandons a field
// at its first failing tag:
//
//   - on a scalar field, the tags AFTER it do not run;
//   - on a nested struct, validation does not descend into it.
//
// Put it first in the tag list. Parameters are pairs of (field name, value).
//
// The motivating case is a cross-field comparison whose right-hand side carries
// a sentinel. `ltefield=MaxOpenConns` reads as "at most MaxOpenConns", but when
// MaxOpenConns is 0 meaning UNLIMITED it is not an upper bound at all, and the
// tag rejects a perfectly good config:
//
//	MaxIdleConns int `validate:"stop_if=MaxOpenConns 0,ltefield=MaxOpenConns"`
//	MaxOpenConns int // 0 = unlimited
//
// Not to be confused with validator's built-in "skip_unless", which despite its
// name never skips anything: it returns hasValue(fl), a presence check in the
// required_* family, and the tags after it still run. No built-in stops
// validation the way these do, which is why they exist.
//
// The names deliberately stay out of the upstream "skip_*" namespace.
// RegisterValidationCtx silently REPLACES a built-in of the same name and
// returns nil, so a collision would change behaviour for every consumer with
// nothing to announce it.
//
// Panics if the number of parameters is not even.
func stopIfImpl(_ context.Context, fl validator.FieldLevel) bool {
	params := parseOneOfParam2(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for %s %s", fl.GetTag(), fl.FieldName()))
	}
	for i := 0; i < len(params); i += 2 {
		// A missing field is not a match, so a typo'd field name never silently
		// disables the rules that follow.
		if requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return false
		}
	}
	return true
}

func skipNestedUnlessWrapper(next ValidatorFunc) ValidatorFunc {
	return func(ctx context.Context, v any) error {
		err := next(ctx, v)
		if err == nil {
			return nil
		}
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			filtered := lo.Filter(verr, func(e validator.FieldError, _ int) bool {
				return !lo.Contains(stopTags, e.Tag())
			})
			if len(filtered) == 0 {
				return nil
			}
			return filtered
		}
		return err
	}
}

// ValidatorWithSkipNestedUnless wraps a validator with support for the
// conditional "stop" tags, which halt validation of a field based on the values
// of other fields in the same struct.
//
// The wrapper performs two functions:
//  1. Registers "stop_if", "stop_unless", and "skip_nested_unless" (a
//     deprecated alias of stop_unless, kept so existing tags keep working)
//  2. Filters out their errors, which mean "stop validating here", not "this
//     value is invalid"
//
// The name is historical — it predates stop_if/stop_unless — and is kept
// because it is part of the public API.
//
// Parameters:
//   - validator: The base validator to wrap
//
// Returns:
//   - Validator: A wrapped validator supporting the stop tags
//
// Panics if registration of any tag fails
func ValidatorWithSkipNestedUnless(validator Validator) Validator {
	for tag, impl := range stopTagImpls {
		if err := validator.RegisterValidationCtx(tag, impl); err != nil {
			panic(fmt.Sprintf("failed to register validation %q: %v", tag, err))
		}
	}
	return &wrappedValidator{
		Validator:     validator,
		structCtxFunc: skipNestedUnlessWrapper(validator.StructCtx),
	}
}
