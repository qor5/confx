package confx_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/qor5/confx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToSliceHookFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		to       any
		expected any
		wantErr  string
	}{
		{
			name:     "bool slice",
			input:    "true,false,true",
			to:       []bool{},
			expected: []bool{true, false, true},
			wantErr:  "",
		},
		{
			name:     "float32 slice",
			input:    "1.1,2.2,3.3",
			to:       []float32{},
			expected: []float32{1.1, 2.2, 3.3},
			wantErr:  "",
		},
		{
			name:     "float64 slice",
			input:    "1.1,2.2,3.3",
			to:       []float64{},
			expected: []float64{1.1, 2.2, 3.3},
			wantErr:  "",
		},
		{
			name:     "int slice",
			input:    "1,2,3",
			to:       []int{},
			expected: []int{1, 2, 3},
			wantErr:  "",
		},
		{
			name:     "int32 slice",
			input:    "1,2,3",
			to:       []int32{},
			expected: []int32{1, 2, 3},
			wantErr:  "",
		},
		{
			name:     "int64 slice",
			input:    "1,2,3",
			to:       []int64{},
			expected: []int64{1, 2, 3},
			wantErr:  "",
		},
		{
			name:  "duration slice",
			input: "1s,2s,3s",
			to:    []time.Duration{},
			expected: []time.Duration{
				1 * time.Second,
				2 * time.Second,
				3 * time.Second,
			},
			wantErr: "",
		},
		{
			name:     "string slice without spaces",
			input:    "a,b,c",
			to:       []string{},
			expected: []string{"a", "b", "c"},
			wantErr:  "",
		},
		{
			name:     "string slice with spaces",
			input:    "a,b ,c",
			to:       []string{},
			expected: []string{"a", "b ", "c"},
			wantErr:  "",
		},
		{
			name:     "empty string slice",
			input:    " ",
			to:       []string{},
			expected: []string{" "},
			wantErr:  "",
		},
		{
			name:     "uint slice",
			input:    "1,2,3",
			to:       []uint{},
			expected: []uint{1, 2, 3},
			wantErr:  "",
		},
		{
			name:     "uint32 slice",
			input:    "1,2,3",
			to:       []uint32{},
			expected: []uint32{1, 2, 3},
			wantErr:  "",
		},
		{
			name:     "uint64 slice",
			input:    "1,2,3",
			to:       []uint64{},
			expected: []uint64{1, 2, 3},
			wantErr:  "",
		},
		{
			name:     "byte slice",
			input:    base64.StdEncoding.EncodeToString([]byte("abc")),
			to:       []byte{},
			expected: []byte("abc"),
			wantErr:  "",
		},
		{
			name:     "empty int slice",
			input:    "",
			to:       []int{},
			expected: []int{},
			wantErr:  "",
		},
		{
			name:     "invalid int8 slice",
			input:    "1,2,3",
			to:       []int8{}, // because flag set does not support []int8 , keep consistent
			expected: nil,
			wantErr:  `unsupported slice element type "int8"`,
		},
		{
			name:     "invalid uint8 slice",
			input:    "1,2,3", // []uint8 == []byte , so this should be a base64 encoded string
			to:       []uint8{},
			expected: nil,
			wantErr:  "illegal base64 data at input byte",
		},
		{
			name:     "invalid int in slice",
			input:    "1,abc,3",
			to:       []int{},
			expected: nil,
			wantErr:  `invalid value "abc"`,
		},
		{
			name:     "invalid bool in slice",
			input:    "true,invalid,false",
			to:       []bool{},
			expected: nil,
			wantErr:  `invalid value "invalid"`,
		},
		{
			name:     "invalid float in slice",
			input:    "1.1,abc,3.3",
			to:       []float64{},
			expected: nil,
			wantErr:  `invalid value "abc"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := &mapstructure.DecoderConfig{
				DecodeHook: confx.StringToSliceHookFunc(","),
				Result:     &tt.to,
			}
			d, err := mapstructure.NewDecoder(decoder)
			require.NoError(t, err)

			err = d.Decode(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.to)
			}
		})
	}
}

func TestStringToMapHookFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		to       any
		expected any
		wantErr  string
	}{
		{
			name:     "empty input",
			input:    "",
			to:       map[string]string{},
			expected: map[string]string{},
			wantErr:  "",
		},
		{
			name:     "string to int",
			input:    "a=1,b=2,c=3",
			to:       map[string]int{},
			expected: map[string]int{"a": 1, "b": 2, "c": 3},
			wantErr:  "",
		},
		{
			name:     "string to int64",
			input:    "a=1,b=2,c=3",
			to:       map[string]int64{},
			expected: map[string]int64{"a": 1, "b": 2, "c": 3},
			wantErr:  "",
		},
		{
			name:     "string to string",
			input:    "a=1,b=2,c=3",
			to:       map[string]string{},
			expected: map[string]string{"a": "1", "b": "2", "c": "3"},
			wantErr:  "",
		},
		{
			name:     "invalid string to bool",
			input:    "a=true,b=false,c=true", // because flag set does not support map[string]bool , keep consistent
			to:       map[string]bool{},
			expected: nil,
			wantErr:  `unsupported map value type "bool"`,
		},
		{
			name:     "invalid float64 key",
			input:    "1.1=a,2.2=b,3.3=c",
			to:       map[float64]string{},
			expected: nil,
			wantErr:  `only string keys are supported for map type, got "float64"`,
		},
		{
			name:     "invalid key-value pair",
			input:    "a=1,b=2,c",
			to:       map[string]int{},
			expected: nil,
			wantErr:  `invalid key-value pair "c"`,
		},
		{
			name:     "invalid key-value separator",
			input:    "a:1,b:2,c:3",
			to:       map[string]int{},
			expected: nil,
			wantErr:  `invalid key-value pair "a:1"`,
		},
		{
			name:     "invalid int value",
			input:    "a=1,b=abc,c=3",
			to:       map[string]int{},
			expected: nil,
			wantErr:  `invalid value "abc" for key "b"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := &mapstructure.DecoderConfig{
				DecodeHook: confx.StringToMapHookFunc(",", "="),
				Result:     &tt.to,
			}
			d, err := mapstructure.NewDecoder(decoder)
			require.NoError(t, err)

			err = d.Decode(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.to)
			}
		})
	}
}

func TestRead(t *testing.T) {
	expectedTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	configFileContent := `
port: 6060
logFiles:
  - "/var/log/app3.log"
  - "/var/log/app4.log"
retryCounts:
  - 33
  - 55
# verbose: true
MaxIdleConns: 10
timeout: "33s"
database:
  host: "localhost"
  port: 5433
extra:
  stringSlice: "a,b,c"
  intSlice: "1,2"
  time: "2023-01-02T15:04:05Z"
ignore: "ignored" # will be ignored
`
	conf, err := confx.Read[*TestConfig]("yaml", strings.NewReader(configFileContent))
	require.NoError(t, err)

	expected := &TestConfig{
		Port:         6060,
		LogFiles:     []string{"/var/log/app3.log", "/var/log/app4.log"},
		RetryCounts:  []int{33, 55},
		Verbose:      false,
		MaxIdleConns: 10,
		Timeout:      33 * time.Second,
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5433,
		},
		Extra: ExtraConfig{
			StringSlice: []string{"a", "b", "c"},
			IntSlice:    []int{1, 2},
			Time:        expectedTime,
		},
		Ignore: "",
	}
	assert.Equal(t, expected, conf)

	{
		type Config struct {
			Host **string `confx:"host"`
		}
		conf, err := confx.Read[*Config]("yaml", strings.NewReader(`host: "localhost"`))
		require.NoError(t, err)
		assert.Equal(t, "localhost", **conf.Host)

		conf, err = confx.Read[*Config]("yaml", strings.NewReader(``))
		require.NoError(t, err)
		assert.Nil(t, conf.Host)
	}
}

func TestReadErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		typ     string
	}{
		{
			name: "invalid yaml",
			content: `
port: "invalid"
logFiles:
  - /var/log/app1.log
`,
			typ: "yaml",
		},
		{
			name: "invalid json",
			content: `{
	"port": "invalid",
	"logFiles": ["/var/log/app1.log"]
}`,
			typ: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := confx.Read[TestConfig](tt.typ, strings.NewReader(tt.content))
			assert.Error(t, err)
		})
	}
}
