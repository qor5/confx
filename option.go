package confx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Option func(opts *initOptions)

// DefaultTagName is the default tag name for struct fields
var DefaultTagName = "confx"

// DefaultUsageTagName is the default tag name for field usage descriptions
var DefaultUsageTagName = "usage"

type initOptions struct {
	flagSet       *pflag.FlagSet
	envPrefix     string
	tagName       string
	usageTagName  string
	viperInstance *viper.Viper
	validator     Validator
	fieldHook     func(f *Field) (*Field, error)
}

// WithFlagSet sets a custom pflag.FlagSet instance for parsing command line flags
// If not set, the default pflag.CommandLine is used.
func WithFlagSet(flagSet *pflag.FlagSet) Option {
	if flagSet == nil {
		panic("flagSet cannot be nil")
	}
	return func(opts *initOptions) {
		opts.flagSet = flagSet
	}
}

// WithEnvPrefix sets a custom environment variable prefix for reading configuration from environment variables.
// If not set, environment variables are read without a prefix.
func WithEnvPrefix(envPrefix string) Option {
	if envPrefix == "" {
		panic("envPrefix cannot be empty")
	}
	return func(opts *initOptions) {
		opts.envPrefix = envPrefix
	}
}

type Field struct {
	ViperKey string
	FlagKey  string
	EnvKey   string
	Usage    string
}

// WithFieldHook sets a custom field hook function that maps configuration field names to Viper keys, flag names, environment variable names and usage strings.
func WithFieldHook(hook func(f *Field) (*Field, error)) Option {
	if hook == nil {
		panic("hook cannot be nil")
	}
	return func(opts *initOptions) {
		opts.fieldHook = hook
	}
}

// WithValidator sets a custom validator instance for validating the configuration struct.
// If not set, the default ValidatorWithSkipNestedUnless(validator.New(validator.WithRequiredStructEnabled())) is used.
func WithValidator(v Validator) Option {
	if v == nil {
		panic("validator cannot be nil")
	}
	return func(opts *initOptions) {
		opts.validator = v
	}
}

// WithViper sets a custom Viper instance for reading configuration from different sources.
// If not set, the default viper.GetViper() is used.
func WithViper(v *viper.Viper) Option {
	if v == nil {
		panic("viperInstance cannot be nil")
	}
	return func(opts *initOptions) {
		opts.viperInstance = v
	}
}

// WithTagName sets a custom struct tag name for reading configuration from struct fields.
// If not set, the default tag name is "confx".
func WithTagName(tagName string) Option {
	if tagName == "" {
		panic("tagName cannot be empty")
	}
	return func(opts *initOptions) {
		opts.tagName = tagName
	}
}

// WithUsageTagName sets a custom struct tag name for reading field usage descriptions.
// If not set, the default tag name is "usage".
func WithUsageTagName(usageTagName string) Option {
	if usageTagName == "" {
		panic("usageTagName cannot be empty")
	}
	return func(opts *initOptions) {
		opts.usageTagName = usageTagName
	}
}
