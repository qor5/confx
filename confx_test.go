package confx_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qor5/confx"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DatabaseConfig struct {
	Host string `confx:"host" usage:"Database host" validate:"required"`
	Port int    `confx:"port" usage:"Database port" validate:"gte=1,lte=65535"`
}

type JWTConfig struct {
	SigningAlgorithm string `confx:"signingAlgorithm"`
	PrivateKey       string `confx:"privateKey"`
	PublicKey        string `confx:"publicKey"`
}

type AuthConfig struct {
	JWT                *JWTConfig    `confx:"jwt"`
	AccessTokenMaxAge  time.Duration `confx:"accessTokenMaxAge"`
	RefreshTokenMaxAge time.Duration `confx:"refreshTokenMaxAge"`
}

type JSONPath struct {
	Path string `confx:"path" json:"path"`
	Hash bool   `confx:"hash" json:"hash"`
}

type ProtectionConfig struct {
	PrivateFields []JSONPath `confx:"privateFields"`
	PublicFields  []JSONPath `confx:"publicFields"`
}

type ExtraConfig struct {
	Bool               bool              `confx:"bool"`
	Float32            float32           `confx:"float32"`
	Float64            float64           `confx:"float64"`
	Int                int               `confx:"int"`
	Int8               int8              `confx:"int8"`
	Int16              int16             `confx:"int16"`
	Int32              int32             `confx:"int32"`
	Int64              int64             `confx:"int64"`
	Duration           time.Duration     `confx:"duration"`
	String             string            `confx:"string"`
	Uint64             uint64            `confx:"uint64"`
	Uint32             uint32            `confx:"uint32"`
	Uint16             uint16            `confx:"uint16"`
	Uint8              uint8             `confx:"uint8"`
	Uint               uint              `confx:"uint"`
	BoolSlice          []bool            `confx:"boolSlice" validate:"required,min=1"`
	Float32Slice       []float32         `confx:"float32Slice"`
	Float64Slice       []float64         `confx:"float64Slice"`
	IntSlice           []int             `confx:"intSlice"`
	Int32Slice         []int32           `confx:"int32Slice"`
	Int64Slice         []int64           `confx:"int64Slice"`
	DurationSlice      []time.Duration   `confx:"durationSlice"`
	StringSlice        []string          `confx:"stringSlice"`
	UintSlice          []uint            `confx:"uintSlice"`
	Uint32Slice        []uint32          `confx:"uint32Slice"`
	Uint64Slice        []uint64          `confx:"uint64Slice"`
	Uint8Slice         []uint8           `confx:"uint8Slice"`
	Bytes              []byte            `confx:"bytes"`
	StringToString     map[string]string `confx:"stringToString"`
	StringToInt        map[string]int    `confx:"stringToInt"`
	StringToInt64      map[string]int64  `confx:"stringToInt64"`
	EmptyStringToInt64 map[string]int64  `confx:"emptyStringToInt64"`
	BytesForEnv        []byte            `confx:"bytesForEnv"`
	Time               time.Time         `confx:"time" validate:"required"`
	StringPtr          *string           `confx:"stringPtr"`
	StringPtrPtr       **string          `confx:"stringPtrPtr"`
}

// TestConfig defines a configuration structure for testing purposes.
type TestConfig struct {
	Port         int              `confx:"port" usage:"Port to run the server on" validate:"gte=1,lte=65535"`
	LogFiles     []string         `confx:"logFiles" usage:"List of log files" validate:"dive,required"`
	RetryCounts  []int            `confx:"retryCounts" usage:"List of retry counts" validate:"dive,gte=0"`
	Verbose      bool             `confx:"verbose" usage:"Enable verbose logging"`
	MaxIdleConns int              `usage:"Maximum number of idle connections"` // Missing confx tag, should default to field name.
	Timeout      time.Duration    `confx:"timeout" usage:"Request timeout duration" validate:"gte=0"`
	Database     DatabaseConfig   `confx:"database"`
	Auth         *AuthConfig      `confx:"auth"`
	Extra        ExtraConfig      `confx:"extra"`
	Protection   ProtectionConfig `confx:"protection"`
	Ignore       string           `confx:"-"` // Ignored field.
	private      string           // Private field.
	privateFunc  func()           // Private function.
}

func TestInitializeLoader(t *testing.T) {
	viper.Reset()

	now := time.Now()
	def := TestConfig{
		Port:         8080,
		LogFiles:     []string{"/var/log/app1.log", "/var/log/app2.log"},
		RetryCounts:  []int{3, 5},
		Verbose:      true,
		MaxIdleConns: 1,
		Timeout:      30 * time.Second,
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Extra: ExtraConfig{
			Bool:          true,
			Float32:       1.1,
			Float64:       2.2,
			Int:           3,
			Int8:          4,
			Int16:         5,
			Int32:         6,
			Int64:         7,
			Duration:      8 * time.Second,
			String:        "string",
			Uint64:        9,
			Uint32:        10,
			Uint16:        11,
			Uint8:         12,
			Uint:          13,
			BoolSlice:     []bool{true, false},
			Float32Slice:  []float32{1.1, 2.2},
			Float64Slice:  []float64{1.1, 2.2},
			IntSlice:      []int{1, 2},
			Int32Slice:    []int32{1, 2},
			Int64Slice:    []int64{1, 2},
			DurationSlice: []time.Duration{1 * time.Second, 2 * time.Second},
			StringSlice:   []string{"a", "b", "c"},
			UintSlice:     []uint{1, 2},
			Uint32Slice:   []uint32{1, 2},
			Uint64Slice:   []uint64{1, 2},
			Uint8Slice:    []uint8{1, 2},
			Bytes:         []byte("abc"),
			Time:          now,
			StringToInt64: map[string]int64{
				"keyA": 1,
			},
		},
		Ignore:      "ignored",
		private:     "private",
		privateFunc: func() {},
	}

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	require.NoError(t, err)

	// Create a temporary configuration file.
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "config.yaml")
	configFileContent := `
port: 6060
logFiles:
  - "/var/log/app3.log"
  - "/var/log/app4.log"
retryCounts:
  - 33
  - 55
# verbose: false
MaxIdleConns: 10
timeout: "33s"
database:
  host: "localhost"
  port: 5433
ignore: "ignored" # will be ignored
protection:
  privateFields:
    - path: fieldA
      hash: true
    - path: fieldB
      hash: true
  publicFields:
    - path: fieldC
      hash: true
    - path: fieldD
      hash: true
`
	err = os.WriteFile(configFilePath, []byte(configFileContent), 0o644)
	require.NoError(t, err)

	// Set environment variables.
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_LOG_FILES", "/tmp/app1.log,/tmp/app2.log")
	t.Setenv("APP_DATABASE_HOST", "127.0.0.1")
	t.Setenv("APP_DATABASE_PORT", "3306")
	t.Setenv("APP_EXTRA_STRING_SLICE", "X,Y,Z")
	t.Setenv("APP_EXTRA_INT_SLICE", "888,999")
	t.Setenv("APP_EXTRA_STRING_TO_STRING", "key1=value1,key2=value2")
	t.Setenv("APP_EXTRA_BYTES_FOR_ENV", base64.StdEncoding.EncodeToString([]byte("xyz")))
	t.Setenv("APP_PROTECTION_PUBLIC_FIELDS", `[{"path":"fieldCFromEnv","hash":false},{"path":"fieldDFromEnv","hash":false}]`)

	// Simulate command-line arguments.
	args := []string{
		"--port=7070",
		"--retry-counts=4", "--retry-counts=6",
		"--log-files=/tmp/app3.log", "--log-files=/tmp/app4.log",
		"--extra-bool-slice=false", "--extra-bool-slice=true", "--extra-bool-slice=false",
		"--extra-float64-slice=11.11,22.22",
		"--extra-string-to-int=key1=1",
		"--extra-string-to-int=key2=2",
		"--extra-string-to-int64=key1=1,key2=2",
		`--protection-private-fields=[{"path":"fieldAFromFlag","hash":false},{"path":"fieldBFromFlag","hash":false}]`,
	}
	err = flagSet.Parse(args)
	require.NoError(t, err)

	// Load configuration with the config file path.
	config, err := loader(context.Background(), configFilePath)
	require.NoError(t, err)

	expected := TestConfig{
		Port:         7070,                                       // flag.
		LogFiles:     []string{"/tmp/app3.log", "/tmp/app4.log"}, // flag.
		RetryCounts:  []int{4, 6},                                // flag.
		Verbose:      true,                                       // default.
		MaxIdleConns: 10,                                         // yaml.
		Timeout:      33 * time.Second,                           // yaml.
		Database: DatabaseConfig{
			Host: "127.0.0.1", // env.
			Port: 3306,        // env.
		},
		Auth: &AuthConfig{
			JWT: &JWTConfig{},
		},
		Protection: ProtectionConfig{
			PrivateFields: []JSONPath{
				{
					Path: "fieldAFromFlag", // flag
					Hash: false,
				},
				{
					Path: "fieldBFromFlag", // flag
					Hash: false,
				},
			},
			PublicFields: []JSONPath{
				{
					Path: "fieldCFromEnv", // env
					Hash: false,
				},
				{
					Path: "fieldDFromEnv", // env
					Hash: false,
				},
			},
		},
		Extra: ExtraConfig{
			Bool:          true,
			Float32:       1.1,
			Float64:       2.2,
			Int:           3,
			Int8:          4,
			Int16:         5,
			Int32:         6,
			Int64:         7,
			Duration:      8 * time.Second,
			String:        "string",
			Uint64:        9,
			Uint32:        10,
			Uint16:        11,
			Uint8:         12,
			Uint:          13,
			BoolSlice:     []bool{false, true, false}, // flag
			Float32Slice:  []float32{1.1, 2.2},
			Float64Slice:  []float64{11.11, 22.22}, // flag
			IntSlice:      []int{888, 999},         // env
			Int32Slice:    []int32{1, 2},
			Int64Slice:    []int64{1, 2},
			DurationSlice: []time.Duration{1 * time.Second, 2 * time.Second},
			StringSlice:   []string{"X", "Y", "Z"}, // env
			UintSlice:     []uint{1, 2},
			Uint32Slice:   []uint32{1, 2},
			Uint64Slice:   []uint64{1, 2},
			Uint8Slice:    []uint8{1, 2},
			Bytes:         []byte("abc"),
			StringToString: map[string]string{
				"key1": "value1",
				"key2": "value2",
			}, // env
			StringToInt: map[string]int{
				"key1": 1,
				"key2": 2,
			}, // flag
			StringToInt64: map[string]int64{
				"key1": 1,
				"key2": 2,
			}, // flag
			EmptyStringToInt64: map[string]int64{},
			BytesForEnv:        []byte("xyz"),
			Time:               now,
			StringPtr:          lo.ToPtr(""),
			StringPtrPtr:       lo.ToPtr(lo.ToPtr("")),
		},
		Ignore:      "", // Ignored
		private:     "",
		privateFunc: nil,
	}

	// Compare Time fields separately since they might have different internal representations
	assert.WithinDuration(t, expected.Extra.Time, config.Extra.Time, time.Second)

	// Set Time fields to zero value for full struct comparison
	expected.Extra.Time = time.Time{}
	config.Extra.Time = time.Time{}
	assert.Equal(t, expected, config)
}

func TestValidationFailure(t *testing.T) {
	viper.Reset()

	def := &TestConfig{
		Port:         8080,
		LogFiles:     []string{"/var/log/app1.log", ""}, // Invalid: empty string
		RetryCounts:  []int{-1, 5},                      // Invalid: negative number
		Verbose:      false,
		MaxIdleConns: 1,
		Timeout:      -10 * time.Second, // Invalid: negative duration
		Database: DatabaseConfig{
			Host: "",    // Invalid: required
			Port: 70000, // Invalid: exceeds max
		},
		Extra: ExtraConfig{
			Time: time.Time{}, // Invalid: zero value
		},
	}

	flagSet := pflag.NewFlagSet("test_validation_failure", pflag.ContinueOnError)
	loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	require.NoError(t, err)

	// Load configuration without specifying a config file path.
	config, err := loader(context.Background(), "")
	assert.ErrorContains(t, err, `Key: 'TestConfig.LogFiles[1]' Error:Field validation for 'LogFiles[1]' failed on the 'required' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.RetryCounts[0]' Error:Field validation for 'RetryCounts[0]' failed on the 'gte' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.Timeout' Error:Field validation for 'Timeout' failed on the 'gte' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.Database.Host' Error:Field validation for 'Host' failed on the 'required' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.Database.Port' Error:Field validation for 'Port' failed on the 'lte' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.Extra.BoolSlice' Error:Field validation for 'BoolSlice' failed on the 'min' tag`)
	assert.ErrorContains(t, err, `Key: 'TestConfig.Extra.Time' Error:Field validation for 'Time' failed on the 'required' tag`)
	assert.Nil(t, config)
}

func TestMapHandling(t *testing.T) {
	viper.Reset()

	now := time.Now()
	def := TestConfig{
		Port: 8080,
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Extra: ExtraConfig{
			BoolSlice:      []bool{true, false},
			StringToString: map[string]string{"key1": "value1"},
			StringToInt:    map[string]int{"key1": 1},
			Time:           now,
		},
	}

	flagSet := pflag.NewFlagSet("test_maps", pflag.ContinueOnError)
	loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	require.NoError(t, err)

	// Test command line arguments for maps
	args := []string{
		"--extra-string-to-string=key2=value2",
		"--extra-string-to-string=key3=value3",
		"--extra-string-to-int=key2=2,key3=3",
	}
	err = flagSet.Parse(args)
	require.NoError(t, err)

	config, err := loader(context.Background(), "")
	require.NoError(t, err)

	expected := map[string]string{
		"key2": "value2",
		"key3": "value3",
	}
	assert.Equal(t, expected, config.Extra.StringToString)

	expectedInts := map[string]int{
		"key2": 2,
		"key3": 3,
	}
	assert.Equal(t, expectedInts, config.Extra.StringToInt)
}

func TestTimeHandling(t *testing.T) {
	viper.Reset()

	timeStr := "2023-01-02T15:04:05Z"
	expectedTime, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)

	def := TestConfig{
		Port: 8080,
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Extra: ExtraConfig{
			BoolSlice: []bool{true, false},
			StringToString: map[string]string{
				"key1": "value1",
			},
			Time: time.Now(),
		},
	}

	flagSet := pflag.NewFlagSet("test_time", pflag.ContinueOnError)
	loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	require.NoError(t, err)

	// Create config file with time
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "config.yaml")
	configFileContent := fmt.Sprintf(`
port: 8080
database:
  host: "localhost"
  port: 5432
extra:
  time: "%s"
`, timeStr)

	err = os.WriteFile(configFilePath, []byte(configFileContent), 0o644)
	require.NoError(t, err)

	config, err := loader(context.Background(), configFilePath)
	require.NoError(t, err)

	assert.Equal(t, expectedTime.UTC(), config.Extra.Time.UTC())
}

type SquashConfig struct {
	Port    int    `confx:"port" usage:"Port to run the server on" validate:"gte=1,lte=65535"`
	Host    string `confx:"host" usage:"Host to run the server on" validate:"required"`
	Timeout time.Duration
}

type ConfigWithSquash struct {
	Base    SquashConfig `confx:",squash"`
	LogFile string       `confx:"logFile" usage:"Log file path"`
}

type InvalidSquashConfig struct {
	Time    time.Time `confx:",squash"` // Invalid: time.Time
	LogFile string    `confx:"logFile"`
}

type InvalidSquashConfig2 struct {
	IntField int    `confx:",squash"` // Invalid: non-struct
	LogFile  string `confx:"logFile"`
}

func TestSquashTag(t *testing.T) {
	t.Run("valid squash", func(t *testing.T) {
		viper.Reset()

		def := ConfigWithSquash{
			Base: SquashConfig{
				Port:    8080,
				Host:    "localhost",
				Timeout: 30 * time.Second,
			},
			LogFile: "/var/log/app.log",
		}

		flagSet := pflag.NewFlagSet("test_squash", pflag.ContinueOnError)
		loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
		require.NoError(t, err)

		// Create a temporary configuration file
		tempDir := t.TempDir()
		configFilePath := filepath.Join(tempDir, "config.yaml")
		configFileContent := `
port: 9090
host: "127.0.0.1"
timeout: "60s"
logFile: "/tmp/app.log"
`
		err = os.WriteFile(configFilePath, []byte(configFileContent), 0o644)
		require.NoError(t, err)

		config, err := loader(context.Background(), configFilePath)
		require.NoError(t, err)

		expected := ConfigWithSquash{
			Base: SquashConfig{
				Port:    9090,
				Host:    "127.0.0.1",
				Timeout: 60 * time.Second,
			},
			LogFile: "/tmp/app.log",
		}
		assert.Equal(t, expected, config)
	})

	t.Run("invalid squash time.Time", func(t *testing.T) {
		viper.Reset()

		def := InvalidSquashConfig{
			Time:    time.Now(),
			LogFile: "/var/log/app.log",
		}

		flagSet := pflag.NewFlagSet("test_invalid_squash", pflag.ContinueOnError)
		_, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unsupported squash type: "time.Time"`)
	})

	t.Run("invalid squash non-struct", func(t *testing.T) {
		viper.Reset()

		def := InvalidSquashConfig2{
			IntField: 123,
			LogFile:  "/var/log/app.log",
		}

		flagSet := pflag.NewFlagSet("test_invalid_squash2", pflag.ContinueOnError)
		_, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unsupported squash type: "int"`)
	})
}

type Stringer string

func (s Stringer) String() string {
	return string(s)
}

func TestUnsupportType(t *testing.T) {
	viper.Reset()

	loader, err := confx.Initialize[fmt.Stringer](Stringer("a stringer"))
	require.ErrorContains(t, err, `unsupported type "confx_test.Stringer" (string)`)
	require.Nil(t, loader)

	tests := []struct {
		name    string
		def     any
		wantErr string
	}{
		{
			name: "unsupported chan type",
			def: struct {
				Chan chan int `confx:"chan"`
			}{
				Chan: make(chan int),
			},
			wantErr: `unsupported field type "chan int" (chan) for key "chan"`,
		},
		{
			name: "unsupported func type",
			def: struct {
				Func func() `confx:"func"`
			}{
				Func: func() {},
			},
			wantErr: `unsupported field type "func()" (func) for key "func"`,
		},
		{
			name: "unsupported complex64 type",
			def: struct {
				Complex complex64 `confx:"complex"`
			}{
				Complex: complex(1, 2),
			},
			wantErr: `unsupported field type "complex64" (complex64) for key "complex"`,
		},
		{
			name: "unsupported interface type",
			def: struct {
				Stringer fmt.Stringer `confx:"stringer"`
			}{
				Stringer: Stringer("a stringer"),
			},
			wantErr: `unsupported field type "fmt.Stringer" (interface) for key "stringer"`,
		},
		{
			name: "unsupported slice of func type",
			def: struct {
				Funcs []func() `confx:"funcs"`
			}{
				Funcs: []func(){func() {}},
			},
			wantErr: `flag key "funcs": unsupported slice element type: "func()"`,
		},
		{
			name: "unsupported map key of int type",
			def: struct {
				IntToString map[int]string `confx:"intToString"`
			}{
				IntToString: map[int]string{1: "one"},
			},
			wantErr: `flag key "int-to-string": only string keys are supported for map type, but got "int"`,
		},
		{
			name: "unsupported map value of func type",
			def: struct {
				StringToFunc map[string]func() `confx:"stringToFunc"`
			}{
				StringToFunc: map[string]func(){"key": func() {}},
			},
			wantErr: `flag key "string-to-func": unsupported map value type "func()"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			loader, err := confx.Initialize(tt.def)
			require.ErrorContains(t, err, tt.wantErr)
			require.Nil(t, loader)
		})
	}
}

func TestUint8Slice(t *testing.T) {
	viper.Reset()

	def := struct {
		Bytes []uint8 `confx:"bytes"` // []uint8 == []byte
	}{
		Bytes: []uint8("abc"),
	}

	flagSet := pflag.NewFlagSet("test_uint8_slice", pflag.ContinueOnError)
	loader, err := confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"))
	require.NoError(t, err)

	{
		args := []string{
			"--bytes=x,y,z",
		}
		err = flagSet.Parse(args)
		require.ErrorContains(t, err, "illegal base64 data at input byte")

		conf, err := loader(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, struct {
			Bytes []uint8 `confx:"bytes"`
		}{
			Bytes: []uint8("abc"),
		}, conf)
	}

	{
		args := []string{
			"--bytes=" + base64.StdEncoding.EncodeToString([]byte("xyz")),
		}
		err = flagSet.Parse(args)
		require.NoError(t, err)

		conf, err := loader(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, struct {
			Bytes []uint8 `confx:"bytes"`
		}{
			Bytes: []uint8("xyz"),
		}, conf)
	}
}

func TestWithFieldHook(t *testing.T) {
	viper.Reset()

	def := &TestConfig{
		Port:         8080,
		LogFiles:     []string{"/var/log/app1.log", "/var/log/app2.log"},
		RetryCounts:  []int{3, 5},
		Verbose:      true,
		MaxIdleConns: 1,
		Timeout:      30 * time.Second,
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		},
		Extra: ExtraConfig{
			BoolSlice: []bool{true, false},
			Time:      time.Now(),
		},
	}

	require.Panics(t, func() {
		_, _ = confx.Initialize(def, confx.WithFieldHook(nil))
	})

	{
		flagSet := pflag.NewFlagSet("test_field_hook", pflag.ContinueOnError)
		loader, err := confx.Initialize(
			def,
			confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"),
			confx.WithFieldHook(func(_ *confx.Field) (*confx.Field, error) {
				return nil, fmt.Errorf("hook error")
			}),
		)
		require.ErrorContains(t, err, "hook error")
		require.Nil(t, loader)
	}

	flagSet := pflag.NewFlagSet("test_field_hook", pflag.ContinueOnError)
	loader, err := confx.Initialize(
		def,
		confx.WithFlagSet(flagSet), confx.WithEnvPrefix("APP_"),
		confx.WithFieldHook(func(f *confx.Field) (*confx.Field, error) {
			if f.ViperKey == "port" {
				f.EnvKey = "APP_HTTP_PORT"
			}
			if f.ViperKey == "MaxIdleConns" {
				assert.Equal(t, "max-idle-conns", f.FlagKey)
				assert.Equal(t, "APP_MAX_IDLE_CONNS", f.EnvKey)
				assert.Equal(t, "Maximum number of idle connections", f.Usage)
				// t.Logf("viperKey: %s, flagKey: %s, envKey: %s, usage: %s", f.ViperKey, f.FlagKey, f.EnvKey, f.Usage)
			}
			return f, nil
		}),
	)
	require.NoError(t, err)

	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_HTTP_PORT", "7070")

	config, err := loader(context.Background(), "")
	require.NoError(t, err)

	assert.Equal(t, 7070, config.Port)
}

func TestWithViper(t *testing.T) {
	viper.Reset()

	type Config struct {
		Port int `confx:"port"`
	}

	def := Config{
		Port: 8080,
	}

	require.Panics(t, func() {
		_, _ = confx.Initialize(def, confx.WithViper(nil))
	})

	v := viper.New()

	flagSet := pflag.NewFlagSet("test_with_viper", pflag.ContinueOnError)
	loader, err := confx.Initialize(
		def,
		confx.WithFlagSet(flagSet),
		confx.WithEnvPrefix("APP_"),
		confx.WithViper(v),
	)
	require.NoError(t, err)

	conf, err := loader(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, def, conf)
	require.Equal(t, 8080, v.GetInt("port"))
	require.NotEqual(t, v, viper.GetInt("port"))
}

func TestWithTagName(t *testing.T) {
	viper.Reset()

	type Config struct {
		Port int `confz:"port"`
	}

	def := Config{
		Port: 8080,
	}

	require.Panics(t, func() {
		_, _ = confx.Initialize(def, confx.WithTagName(""))
	})

	flagSet := pflag.NewFlagSet("test_with_tag_name", pflag.ContinueOnError)
	loader, err := confx.Initialize(
		def,
		confx.WithFlagSet(flagSet),
		confx.WithEnvPrefix("APP_"),
		confx.WithTagName("confz"),
	)
	require.NoError(t, err)

	conf, err := loader(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, def, conf)
}
