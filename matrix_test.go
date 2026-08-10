package confx

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type matrixInner struct {
	Name string `validate:"required,min=6"`
}

func newStopValidator() Validator {
	return ValidatorWithSkipNestedUnless(validator.New(validator.WithRequiredStructEnabled()))
}

// The tags are NOT scoped to nested structs or to scalar fields — both stop
// validation of whatever they sit on, because validator abandons a field at its
// first failing tag. This is why they are named stop_* rather than skip_nested_*
// or skip_rest_*: either of those words describes only half of what they do.
func TestStopTagsAreNotScopedToNestedOrScalar(t *testing.T) {
	v := newStopValidator()

	t.Run("stop_if on a scalar stops the tags after it", func(t *testing.T) {
		type S struct {
			X int `validate:"stop_if=Y 0,gte=100"`
			Y int
		}
		assert.NoError(t, v.StructCtx(context.Background(), S{X: 1, Y: 0}))
		assert.Error(t, v.StructCtx(context.Background(), S{X: 1, Y: 9}))
	})

	t.Run("stop_if on a nested struct prevents descending", func(t *testing.T) {
		type S struct {
			In matrixInner `validate:"stop_if=Y 0"`
			Y  int
		}
		assert.NoError(t, v.StructCtx(context.Background(), S{In: matrixInner{Name: "ab"}, Y: 0}))
		assert.Error(t, v.StructCtx(context.Background(), S{In: matrixInner{Name: "ab"}, Y: 9}))
	})

	t.Run("stop_unless on a scalar stops the tags after it", func(t *testing.T) {
		type S struct {
			X int `validate:"stop_unless=Y 1,gte=100"`
			Y int
		}
		assert.NoError(t, v.StructCtx(context.Background(), S{X: 1, Y: 0}))
		assert.Error(t, v.StructCtx(context.Background(), S{X: 1, Y: 1}))
	})

	t.Run("stop_unless on a nested struct prevents descending", func(t *testing.T) {
		type S struct {
			In matrixInner `validate:"stop_unless=Y 1"`
			Y  int
		}
		assert.NoError(t, v.StructCtx(context.Background(), S{In: matrixInner{Name: "ab"}, Y: 0}))
		assert.Error(t, v.StructCtx(context.Background(), S{In: matrixInner{Name: "ab"}, Y: 1}))
	})
}

// skip_nested_unless is a deprecated alias of stop_unless and must behave
// identically, including on scalar fields it was never documented for.
func TestSkipNestedUnlessIsAnAliasOfStopUnless(t *testing.T) {
	v := newStopValidator()

	type Old struct {
		X int `validate:"skip_nested_unless=Y 1,gte=100"`
		Y int
	}
	type New struct {
		X int `validate:"stop_unless=Y 1,gte=100"`
		Y int
	}
	for _, y := range []int{0, 1} {
		oldErr := v.StructCtx(context.Background(), Old{X: 1, Y: y})
		newErr := v.StructCtx(context.Background(), New{X: 1, Y: y})
		assert.Equal(t, oldErr == nil, newErr == nil, "Y=%d", y)
	}
}
