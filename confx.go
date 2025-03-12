package confx

import (
	"context"
	"encoding/json"
	"go/ast"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/huandu/go-clone"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Loader[T any] func(ctx context.Context, path string) (T, error)

// Initialize sets up configuration binding by automatically registering command-line flags,
// binding environment variables, loading configuration files, and validating the final configuration.
//
// It leverages reflection to traverse the fields of the provided default configuration struct,
// supports various data types including basic types, slices, maps, and nested structs defined separately,
// and integrates with Viper for configuration management and go-playground/validator for validation.
//
// Note: Even if a field in the default configuration is a nil pointer, it will be assigned a zero value
// after loading. This is because flags typically require a default value, which is usually not nil.
// This behavior aligns with standard configuration requirements, ensuring that all fields are initialized
// with usable values rather than nil pointers.
//
// Parameters:
//   - def: The default configuration struct.
//
// Returns:
//   - Loader[T]: A generic loader function that accepts an optional configuration file path.
//     When invoked, it parses the command-line flags, binds them along with environment
//     variables, loads the configuration file if provided, unmarshal the configuration
//     into the struct, and validates it.
//   - error: An error object if initialization fails.
func Initialize[T any](def T, options ...Option) (Loader[T], error) {
	opts := &initOptions{
		flagSet:       nil,
		envPrefix:     "",
		tagName:       DefaultTagName,
		usageTagName:  DefaultUsageTagName,
		viperInstance: viper.GetViper(),
		validator:     validator.New(validator.WithRequiredStructEnabled()),
	}
	for _, opt := range options {
		opt(opts)
	}
	def = clone.Slowly(def).(T)

	var flagConfig string
	if opts.flagSet == nil {
		opts.flagSet = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
		opts.flagSet.SortFlags = false
		opts.flagSet.StringVarP(&flagConfig, "config", "c", "", "Path to configuration file")
	}

	var collectBinds []func() error
	err := initializeRecursive(opts, reflect.ValueOf(def), "", &collectBinds)
	if err != nil {
		return nil, err
	}

	enhancedValidator := ValidatorWithSkipNestedUnless(opts.validator)

	var once sync.Once
	var onceErr error
	return func(ctx context.Context, confPath string) (T, error) {
		once.Do(func() {
			if !opts.flagSet.Parsed() {
				if err := opts.flagSet.Parse(os.Args[1:]); err != nil {
					if errors.Is(err, pflag.ErrHelp) {
						os.Exit(0)
					}
					onceErr = errors.Wrap(err, "failed to parse flags")
					return
				}
			}

			for _, bind := range collectBinds {
				if err := bind(); err != nil {
					onceErr = err
					return
				}
			}
		})
		var zero T
		if onceErr != nil {
			return zero, onceErr
		}

		if confPath == "" {
			confPath = flagConfig
		}

		if confPath != "" {
			opts.viperInstance.SetConfigFile(confPath)
			if err := opts.viperInstance.ReadInConfig(); err != nil {
				return zero, errors.Wrapf(err, "failed to read config %q", confPath)
			}
		}

		var conf T
		if err := opts.viperInstance.Unmarshal(&conf, DecoderConfigOption(opts.tagName)); err != nil {
			return zero, errors.Wrapf(err, "failed to unmarshal config to %T", conf)
		}

		if err := enhancedValidator.StructCtx(ctx, conf); err != nil {
			return zero, errors.Wrap(err, "validation failed for config")
		}

		return conf, nil
	}, nil
}

var envReplacer = strings.NewReplacer(".", "_", "-", "_")

// unwrapOrNew dereferences a pointer type reflect.Value until a non-pointer
// type is reached. If the input is a nil pointer, it initializes a new value
// of the underlying type and returns it. This function is useful for ensuring
// a non-pointer type reflect.Value is returned from a potential nil pointer.
func unwrapOrNew(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem())
		}
		v = v.Elem()
	}
	return v
}

func unwrapType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

var reKebabCaseFixDigital = regexp.MustCompile(`-(\d+)`)

var (
	typeDuration = reflect.TypeOf(time.Duration(5))
	typeTime     = reflect.TypeOf(time.Time{})
)

func initializeRecursive(
	opts *initOptions,
	v reflect.Value,
	parentKey string,
	collectBinds *[]func() error,
) error {
	v = unwrapOrNew(v)
	if v.Kind() != reflect.Struct {
		return errors.Errorf("unsupported type %q (%s)", v.Type().String(), v.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if !ast.IsExported(field.Name) {
			continue
		}
		fieldType := unwrapType(field.Type)
		fieldValue := unwrapOrNew(v.Field(i))
		tag := strings.TrimSpace(field.Tag.Get(opts.tagName))
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = field.Name
		}
		if tag == ",squash" {
			if fieldType.Kind() != reflect.Struct || fieldType == typeTime {
				return errors.Errorf("unsupported squash type: %q", fieldType)
			}
			if err := initializeRecursive(opts, fieldValue, parentKey, collectBinds); err != nil {
				return err
			}
			continue
		}

		viperKey := tag
		if parentKey != "" {
			viperKey = parentKey + "." + viperKey
		}
		flagKey := reKebabCaseFixDigital.ReplaceAllString(lo.KebabCase(viperKey), "${1}")
		envKey := opts.envPrefix + envReplacer.Replace(strings.ToUpper(flagKey))
		usage := strings.TrimSpace(field.Tag.Get(opts.usageTagName))
		if usage == "" {
			usage = viperKey // Fallback to viperKey if usage is not provided.
		}

		if opts.fieldHook != nil {
			f, err := opts.fieldHook(&Field{
				ViperKey: viperKey,
				FlagKey:  flagKey,
				EnvKey:   envKey,
				Usage:    usage,
			})
			if err != nil {
				return err
			}
			viperKey, flagKey, envKey, usage = f.ViperKey, f.FlagKey, f.EnvKey, f.Usage
		}

		switch fieldType.Kind() {
		case reflect.Bool:
			opts.flagSet.Bool(flagKey, fieldValue.Bool(), usage)
		case reflect.Float32:
			opts.flagSet.Float32(flagKey, float32(fieldValue.Float()), usage)
		case reflect.Float64:
			opts.flagSet.Float64(flagKey, fieldValue.Float(), usage)
		case reflect.Int:
			opts.flagSet.Int(flagKey, int(fieldValue.Int()), usage)
		case reflect.Int8:
			opts.flagSet.Int8(flagKey, int8(fieldValue.Int()), usage)
		case reflect.Int16:
			opts.flagSet.Int16(flagKey, int16(fieldValue.Int()), usage)
		case reflect.Int32:
			opts.flagSet.Int32(flagKey, int32(fieldValue.Int()), usage)
		case reflect.Int64:
			if fieldType == typeDuration {
				opts.flagSet.Duration(flagKey, fieldValue.Interface().(time.Duration), usage)
			} else {
				opts.flagSet.Int64(flagKey, fieldValue.Int(), usage)
			}
		case reflect.String:
			opts.flagSet.String(flagKey, fieldValue.String(), usage)
		case reflect.Uint:
			opts.flagSet.Uint(flagKey, uint(fieldValue.Uint()), usage)
		case reflect.Uint8:
			opts.flagSet.Uint8(flagKey, uint8(fieldValue.Uint()), usage)
		case reflect.Uint16:
			opts.flagSet.Uint16(flagKey, uint16(fieldValue.Uint()), usage)
		case reflect.Uint32:
			opts.flagSet.Uint32(flagKey, uint32(fieldValue.Uint()), usage)
		case reflect.Uint64:
			opts.flagSet.Uint64(flagKey, fieldValue.Uint(), usage)
		case reflect.Slice:
			if err := flagSetSlice(opts.flagSet, fieldValue, flagKey, usage); err != nil {
				return err
			}
		case reflect.Map:
			if err := flagSetMap(opts.flagSet, fieldValue, flagKey, usage); err != nil {
				return err
			}
		case reflect.Struct:
			if fieldType == typeTime {
				opts.flagSet.String(flagKey, fieldValue.Interface().(time.Time).Format(time.RFC3339), usage+" (time in RFC3339 format)")
			} else {
				if err := initializeRecursive(opts, fieldValue, viperKey, collectBinds); err != nil {
					return err
				}
				continue
			}
		default:
			return errors.Errorf("unsupported field type %q (%s) for key %q", fieldType, fieldType.Kind(), viperKey)
		}

		*collectBinds = append(*collectBinds, func() error {
			if err := opts.viperInstance.BindPFlag(viperKey, opts.flagSet.Lookup(flagKey)); err != nil {
				return errors.Wrapf(err, "failed to bind flag %q", flagKey)
			}
			return nil
		})

		*collectBinds = append(*collectBinds, func() error {
			if err := opts.viperInstance.BindEnv(viperKey, envKey); err != nil {
				return errors.Wrapf(err, "failed to bind env %q", envKey)
			}
			return nil
		})
	}

	return nil
}

func convertSlice[T any](fieldValue reflect.Value, convertFunc func(reflect.Value) T) []T {
	var slice []T
	for i := 0; i < fieldValue.Len(); i++ {
		slice = append(slice, convertFunc(fieldValue.Index(i)))
	}
	return slice
}

func flagSetSlice(flagSet *pflag.FlagSet, fieldValue reflect.Value, flagKey, usage string) error {
	elemType := fieldValue.Type().Elem()
	switch elemType.Kind() {
	case reflect.Bool:
		flagSet.BoolSlice(flagKey, convertSlice(fieldValue, func(v reflect.Value) bool {
			return v.Bool()
		}), usage)
	case reflect.Float32:
		flagSet.Float32Slice(flagKey, convertSlice(fieldValue, func(v reflect.Value) float32 {
			return float32(v.Float())
		}), usage)
	case reflect.Float64:
		flagSet.Float64Slice(flagKey, convertSlice(fieldValue, func(v reflect.Value) float64 {
			return v.Float()
		}), usage)
	case reflect.Int:
		flagSet.IntSlice(flagKey, convertSlice(fieldValue, func(v reflect.Value) int {
			return int(v.Int())
		}), usage)
	case reflect.Int32:
		flagSet.Int32Slice(flagKey, convertSlice(fieldValue, func(v reflect.Value) int32 {
			return int32(v.Int())
		}), usage)
	case reflect.Int64:
		if elemType == typeDuration {
			flagSet.DurationSlice(flagKey, convertSlice(fieldValue, func(v reflect.Value) time.Duration {
				return time.Duration(v.Int())
			}), usage)
		} else {
			flagSet.Int64Slice(flagKey, convertSlice(fieldValue, func(v reflect.Value) int64 {
				return v.Int()
			}), usage)
		}
	case reflect.String:
		flagSet.StringSlice(flagKey, convertSlice(fieldValue, func(v reflect.Value) string {
			return v.String()
		}), usage)
	case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		if fieldValue.Type() == reflect.TypeOf([]byte{}) {
			flagSet.BytesBase64(flagKey, fieldValue.Bytes(), usage)
		} else {
			flagSet.UintSlice(flagKey, convertSlice(fieldValue, func(v reflect.Value) uint {
				return uint(v.Uint())
			}), usage)
		}
	case reflect.Struct:
		bs, err := json.Marshal(fieldValue.Interface())
		if err != nil {
			return errors.Wrapf(err, "failed to marshal json, key %q", flagKey)
		}
		flagSet.String(flagKey, string(bs), usage)
	default:
		return errors.Errorf("flag key %q: unsupported slice element type: %q", flagKey, elemType)
	}
	return nil
}

func convertMap[T any](fieldValue reflect.Value, convertFunc func(reflect.Value) T) map[string]T {
	m := make(map[string]T)
	for _, key := range fieldValue.MapKeys() {
		m[key.String()] = convertFunc(fieldValue.MapIndex(key))
	}
	return m
}

func flagSetMap(flagSet *pflag.FlagSet, fieldValue reflect.Value, flagKey, usage string) error {
	keyType := fieldValue.Type().Key()
	if keyType.Kind() != reflect.String {
		return errors.Errorf("flag key %q: only string keys are supported for map type, but got %q", flagKey, keyType)
	}

	elemType := fieldValue.Type().Elem()
	switch elemType.Kind() {
	case reflect.Int:
		flagSet.StringToInt(flagKey, convertMap(fieldValue, func(v reflect.Value) int {
			return int(v.Int())
		}), usage)
	case reflect.Int64:
		flagSet.StringToInt64(flagKey, convertMap(fieldValue, func(v reflect.Value) int64 {
			return v.Int()
		}), usage)
	case reflect.String:
		flagSet.StringToString(flagKey, convertMap(fieldValue, func(v reflect.Value) string {
			return v.String()
		}), usage)
	default:
		return errors.Errorf("flag key %q: unsupported map value type %q", flagKey, elemType)
	}
	return nil
}
