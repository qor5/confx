package confx

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestWithFlagSet(t *testing.T) {
	t.Run("valid flag set", func(t *testing.T) {
		customFlagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
		option := WithFlagSet(customFlagSet)

		opts := &initOptions{}
		option(opts)

		assert.Equal(t, customFlagSet, opts.flagSet)
	})

	t.Run("nil flag set", func(t *testing.T) {
		assert.Panics(t, func() {
			WithFlagSet(nil)
		})
	})
}

func TestWithEnvPrefix(t *testing.T) {
	t.Run("valid env prefix", func(t *testing.T) {
		prefix := "TEST_"
		option := WithEnvPrefix(prefix)

		opts := &initOptions{}
		option(opts)

		assert.Equal(t, prefix, opts.envPrefix)
	})

	t.Run("empty env prefix", func(t *testing.T) {
		assert.Panics(t, func() {
			WithEnvPrefix("")
		})
	})
}

func TestWithValidator(t *testing.T) {
	t.Run("Valid validator", func(t *testing.T) {
		v := validator.New()
		opt := WithValidator(v)

		opts := &initOptions{}
		opt(opts)

		if opts.validator != v {
			t.Errorf("Expected validator to be set, got %v", opts.validator)
		}
	})

	t.Run("Nil validator", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for nil validator, but didn't")
			}
		}()

		WithValidator(nil)
	})
}
