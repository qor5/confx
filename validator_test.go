package confx

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestSkipNestedUnless(t *testing.T) {
	type Foo struct {
		Name string `validate:"required,min=6,max=100"`
	}

	type Bar struct {
		Foo  Foo `validate:"skip_nested_unless=Kind foo Type foo"`
		Kind string
		Type string
	}

	v := ValidatorWithSkipNestedUnless(
		validator.New(validator.WithRequiredStructEnabled()),
	)

	ctx := context.Background()

	// Test basic string matching
	assert.NoError(t, v.StructCtx(ctx, Bar{Foo: Foo{}, Kind: "foo"}))
	assert.NoError(t, v.StructCtx(ctx, Bar{Foo: Foo{}, Kind: "foo", Type: "bar"}))
	assert.ErrorContains(t, v.StructCtx(ctx, Bar{Foo: Foo{}, Kind: "foo", Type: "foo"}), `Key: 'Bar.Foo.Name' Error:Field validation for 'Name' failed on the 'required' tag`)
	assert.NoError(t, v.StructCtx(ctx, Bar{Foo: Foo{Name: "foobar"}, Kind: "foo", Type: "foo"}))

	type NestedTypes struct {
		IntField     int
		UintField    uint
		FloatField   float64
		Float32Field float32
		BoolField    bool
		StringField  string
		SliceField   []string
		PtrField     *string
		Nested       Foo `validate:"skip_nested_unless=IntField 1 UintField 2 FloatField 3.0 Float32Field 1.5 BoolField true StringField foo SliceField 3 PtrField nil"`
	}

	// Test different field types
	strVal := "foo"
	tests := []struct {
		name    string
		input   NestedTypes
		wantErr bool
	}{
		{
			name: "all fields match with valid nested",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: "foobar"},
			},
			wantErr: false,
		},
		{
			name: "all fields match",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: true,
		},
		{
			name: "int field mismatch",
			input: NestedTypes{
				IntField:     2, // should be 1
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "uint field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    3, // should be 2
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "float field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.1, // should be 3.0
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "bool field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    false, // should be true
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "string field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "bar", // should be "foo"
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "slice length mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b"}, // should have length 3
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "ptr field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.5,
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     &strVal, // should be nil
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
		{
			name: "float32 field mismatch",
			input: NestedTypes{
				IntField:     1,
				UintField:    2,
				FloatField:   3.0,
				Float32Field: 1.6, // should be 1.5
				BoolField:    true,
				StringField:  "foo",
				SliceField:   []string{"a", "b", "c"},
				PtrField:     nil,
				Nested:       Foo{Name: ""},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.StructCtx(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseOneOfParam2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple space-separated values",
			input:    "a b c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "values with single quotes",
			input:    "'a b' c",
			expected: []string{"a b", "c"},
		},
		{
			name:     "multiple quoted values",
			input:    "'a b' 'c d'",
			expected: []string{"a b", "c d"},
		},
		{
			name:     "single quoted value",
			input:    "'a b c'",
			expected: []string{"a b c"},
		},
		{
			name:     "mixed quoted and unquoted values",
			input:    "a 'b c' d 'e f'",
			expected: []string{"a", "b c", "d", "e f"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOneOfParam2(tt.input)
			assert.Equal(t, tt.expected, result, "Expected %v, but got %v", tt.expected, result)
		})
	}
}

func TestStopIf(t *testing.T) {
	// The motivating shape: MaxOpenConns == 0 means UNLIMITED, so it is not an
	// upper bound and `ltefield` must not run against it.
	type Pool struct {
		MaxIdleConns int `validate:"stop_if=MaxOpenConns 0,ltefield=MaxOpenConns"`
		MaxOpenConns int
	}

	v := ValidatorWithSkipNestedUnless(
		validator.New(validator.WithRequiredStructEnabled()),
	)

	for _, c := range []struct {
		name       string
		idle, open int
		wantTag    string // "" = no error
	}{
		{"cap set, idle within it", 20, 200, ""},
		{"cap set, idle equals it", 10, 10, ""},
		{"cap set, idle above it", 11, 10, "ltefield"},
		{"unlimited, idle is not compared", 20, 0, ""},
		{"unlimited, both zero", 0, 0, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := v.StructCtx(context.Background(), Pool{c.idle, c.open})
			if c.wantTag == "" {
				assert.NoError(t, err)
				return
			}
			var verr validator.ValidationErrors
			assert.ErrorAs(t, err, &verr)
			assert.Len(t, verr, 1)
			assert.Equal(t, c.wantTag, verr[0].Tag())
		})
	}
}

func TestStopIfDoesNotLeakItsOwnError(t *testing.T) {
	// stop_if works by failing, which stops the tags after it. That failure is
	// an implementation detail and must never reach the caller.
	type S struct {
		A int `validate:"stop_if=B 0,gte=100"`
		B int
	}
	v := ValidatorWithSkipNestedUnless(validator.New(validator.WithRequiredStructEnabled()))

	// B == 0 → skipped, so A = 1 does not have to be >= 100.
	assert.NoError(t, v.StructCtx(context.Background(), S{A: 1, B: 0}))

	// B != 0 → not skipped, so gte=100 applies and reports itself, not stop_if.
	err := v.StructCtx(context.Background(), S{A: 1, B: 7})
	var verr validator.ValidationErrors
	assert.ErrorAs(t, err, &verr)
	assert.Len(t, verr, 1)
	assert.Equal(t, "gte", verr[0].Tag())
}

func TestStopIfMatchesAnyPair(t *testing.T) {
	type S struct {
		A    int `validate:"stop_if=B 0 C 0,gte=100"`
		B, C int
	}
	v := ValidatorWithSkipNestedUnless(validator.New(validator.WithRequiredStructEnabled()))

	assert.NoError(t, v.StructCtx(context.Background(), S{A: 1, B: 0, C: 9}), "B matches")
	assert.NoError(t, v.StructCtx(context.Background(), S{A: 1, B: 9, C: 0}), "C matches")
	assert.Error(t, v.StructCtx(context.Background(), S{A: 1, B: 9, C: 9}), "neither matches")
}

func TestStopIfUnknownFieldDoesNotDisableTheRule(t *testing.T) {
	// A typo in the field name must not silently switch validation off.
	type S struct {
		A int `validate:"stop_if=Nope 0,gte=100"`
		B int
	}
	v := ValidatorWithSkipNestedUnless(validator.New(validator.WithRequiredStructEnabled()))
	assert.Error(t, v.StructCtx(context.Background(), S{A: 1, B: 0}))
}

// Guards the reason stop_if / stop_unless exist at all: validator's built-in
// "skip_unless" does not skip anything despite its name. It returns
// hasValue(fl) — a presence check in the required_* family — so the tags after
// it still run. If upstream ever changes that, this test fails and we can
// reconsider whether stop_if is still needed.
func TestBuiltinSkipUnlessDoesNotActuallySkip(t *testing.T) {
	type S struct {
		A int `validate:"skip_unless=B 0,gte=100"`
		B int
	}
	v := validator.New(validator.WithRequiredStructEnabled())

	// B == 0 matches the condition. If skip_unless skipped, A = 1 would pass.
	err := v.Struct(S{A: 1, B: 0})
	var verr validator.ValidationErrors
	assert.ErrorAs(t, err, &verr)
	assert.Equal(t, "gte", verr[0].Tag(), "built-in skip_unless is documented as a presence check, not a skip")
}
