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

const skipNestedUnlessTag = "skip_nested_unless"

// skipNestedUnless is a validation function that conditionally skips nested struct validation
// based on field values in the parent struct. It is used with the "skip_nested_unless" tag.
//
// The function takes pairs of parameters where each pair consists of:
//  1. A field name to check
//  2. The expected value for that field
//
// If any of the specified field values don't match their expected values, the nested validation
// is skipped by returning false. All pairs must match for validation to proceed.
//
// Example usage in struct tags:
//
//	type Config struct {
//	  Type    string     `validate:"oneof=local remote"`
//	  Local   LocalConf  `validate:"skip_nested_unless=Type local"`
//	  Remote  RemoteConf `validate:"skip_nested_unless=Type remote"`
//	}
//
// In this example:
// - Local config is only validated when Type="local"
// - Remote config is only validated when Type="remote"
//
// Parameters:
//   - ctx: Context (unused)
//   - fl: FieldLevel object providing access to the struct field being validated
//
// Returns:
//   - bool: true if nested validation should proceed, false if it should be skipped
//
// Panics if the number of parameters is not even (must be pairs of field name and expected value)
func skipNestedUnlessImpl(_ context.Context, fl validator.FieldLevel) bool {
	params := parseOneOfParam2(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for skip_nested_unless %s", fl.FieldName()))
	}
	for i := 0; i < len(params); i += 2 {
		// To skip validation, return false to generate the corresponding error, ensuring the nested struct is not validated.
		// The corresponding errors should then be filtered out after the StructCtx method returns.
		// Therefore, this should return false when the condition is not met, preventing further validation.
		if !requireCheckFieldValue(fl, params[i], params[i+1], false) {
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
				return e.Tag() != skipNestedUnlessTag
			})
			if len(filtered) == 0 {
				return nil
			}
			return filtered
		}
		return err
	}
}

// ValidatorWithSkipNestedUnless wraps a validator with support for conditional nested struct validation
// using the "skip_nested_unless" tag. This allows you to skip validation of nested structs based on
// the values of other fields in the parent struct.
//
// The wrapper performs two main functions:
//  1. Registers the "skip_nested_unless" validation tag
//  2. Filters out validation errors from skipped nested structs
//
// Parameters:
//   - validator: The base validator to wrap with skip_nested_unless support
//
// Returns:
//   - Validator: A wrapped validator that supports the skip_nested_unless tag
//
// Panics if registration of the skip_nested_unless validation fails
func ValidatorWithSkipNestedUnless(validator Validator) Validator {
	err := validator.RegisterValidationCtx(skipNestedUnlessTag, skipNestedUnlessImpl)
	if err != nil {
		panic(fmt.Sprintf("failed to register validation: %v", err))
	}
	return &wrappedValidator{
		Validator:     validator,
		structCtxFunc: skipNestedUnlessWrapper(validator.StructCtx),
	}
}
