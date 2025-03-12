package confx

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var DecoderConfigOption = func(tagName string) func(dc *mapstructure.DecoderConfig) {
	return func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			StringToSliceHookFunc(","),
			StringToMapHookFunc(",", "="),
		)
		if tagName != "" {
			dc.TagName = tagName
		} else {
			dc.TagName = DefaultTagName
		}
	}
}

func Read[T any](typ string, r io.Reader) (T, error) {
	return ReadWithTagName[T]("", typ, r)
}

func ReadWithTagName[T any](tagName string, typ string, r io.Reader) (T, error) {
	var zero T

	viperInstance := viper.New()
	viperInstance.SetConfigType(strings.TrimLeft(typ, "."))
	if err := viperInstance.ReadConfig(r); err != nil {
		return zero, errors.Wrap(err, "failed to read config")
	}

	var def T
	if err := viperInstance.Unmarshal(&def, DecoderConfigOption(tagName)); err != nil {
		return zero, errors.Wrap(err, "failed to unmarshal config")
	}

	return def, nil
}

func StringToSliceHookFunc(separator string) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() == reflect.String && to.Kind() == reflect.Slice {
			elemType := to.Elem()
			if unwrapType(elemType).Kind() == reflect.Struct {
				sliceValue := reflect.New(to).Elem()
				err := json.Unmarshal([]byte(data.(string)), sliceValue.Addr().Interface())
				if err != nil {
					return nil, errors.Wrapf(err, "failed to unmarshal json, data: %s", data)
				}
				return sliceValue.Interface(), nil
			}
			str := strings.Trim(data.(string), "[]")
			if to == reflect.TypeOf([]byte{}) {
				return base64.StdEncoding.DecodeString(str)
			}

			parts := strings.Split(str, separator)

			switch elemType.Kind() {
			case reflect.Bool:
				return parseSlice(parts, strconv.ParseBool)
			case reflect.Float32, reflect.Float64:
				return parseSlice(parts, func(s string) (float64, error) {
					return strconv.ParseFloat(s, 64)
				})
			case reflect.Int, reflect.Int32, reflect.Int64:
				if elemType == typeDuration {
					// nolint:gocritic
					return parseSlice(parts, func(s string) (time.Duration, error) {
						return time.ParseDuration(s)
					})
				}
				return parseSlice(parts, func(s string) (int64, error) {
					return strconv.ParseInt(s, 10, 64)
				})
			case reflect.String:
				return parts, nil
			case reflect.Uint, reflect.Uint32, reflect.Uint64:
				return parseSlice(parts, func(s string) (uint64, error) {
					return strconv.ParseUint(s, 10, 64)
				})
			default:
				return nil, errors.Errorf("unsupported slice element type %q", elemType)
			}
		}
		return data, nil
	}
}

func parseSlice[T any](parts []string, parseFunc func(string) (T, error)) ([]T, error) {
	if len(parts) == 0 ||
		(len(parts) == 1 && strings.TrimSpace(parts[0]) == "") {
		return []T{}, nil
	}
	result := make([]T, len(parts))
	for i, part := range parts {
		parsed, err := parseFunc(strings.TrimSpace(part))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid value %q", part)
		}
		result[i] = parsed
	}
	return result, nil
}

func StringToMapHookFunc(separator string, pairSeparator string) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() == reflect.String && to.Kind() == reflect.Map {
			str := strings.Trim(data.(string), "[]")
			if str == "" {
				return reflect.MakeMap(to).Interface(), nil
			}

			pairs := strings.Split(str, separator)

			keyType := to.Key()
			elemType := to.Elem()
			if keyType.Kind() != reflect.String {
				return nil, errors.Errorf("only string keys are supported for map type, got %q", keyType)
			}

			result := reflect.MakeMap(to)
			for _, pair := range pairs {
				kv := strings.SplitN(pair, pairSeparator, 2)
				if len(kv) != 2 {
					return nil, errors.Errorf("invalid key-value pair %q", pair)
				}

				key := kv[0]
				valueStr := kv[1]

				var value any
				var err error

				switch elemType.Kind() {
				case reflect.Int:
					value, err = strconv.Atoi(strings.TrimSpace(valueStr))
				case reflect.Int64:
					value, err = strconv.ParseInt(strings.TrimSpace(valueStr), 10, 64)
				case reflect.String:
					value = valueStr
				default:
					return nil, errors.Errorf("unsupported map value type %q", elemType)
				}

				if err != nil {
					return nil, errors.Wrapf(err, "invalid value %q for key %q", valueStr, key)
				}

				result.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
			}

			return result.Interface(), nil
		}
		return data, nil
	}
}
