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
