package confx

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// ExpectedValidationError represents an expected validation error for testing purposes.
type ExpectedValidationError struct {
	Path string // Path to the field that should fail validation, using dot notation for nested fields
	Tag  string // Expected validation tag that should fail
}

// ExpectedValidation represents the expected validation result for a config.
type ExpectedValidation struct {
	Config         any // The config instance to validate
	Name           string
	ExpectedErrors []ExpectedValidationError // Expected validation errors. if empty, no validation errors are expected.
}

// ValidationSuite provides utilities for testing config validation rules.
type ValidationSuite struct {
	t         *testing.T
	validator Validator
}

// NewValidationSuite creates a new ValidationSuite for testing config validation rules.
//
// Example:
//
//		type MyConfig struct {
//		    Name string `validate:"required"`
//		    Age  int    `validate:"gte=0,lte=150"`
//		}
//
//		func TestMyConfigValidation(t *testing.T) {
//		    suite := confx.NewValidationSuite(t)
//
//		    suite.RunTests([]confx.ValidationExpectation{
//		        {
//					Name: 			"valid config",
//		            Config:         &MyConfig{Name: "John", Age: 30},
//		        },
//		        {
//	             Name:           "invalid config",
//		            Config:         &MyConfig{Age: -1},
//		            ExpectedErrors: []confx.ValidationError{
//		                {Path: "Name", Tag: "required"},
//		                {Path: "Age", Tag: "gte"},
//		            },
//		        },
//		    })
//		}
func NewValidationSuite(t *testing.T) *ValidationSuite {
	return &ValidationSuite{
		t: t,
		validator: ValidatorWithSkipNestedUnless(
			validator.New(validator.WithRequiredStructEnabled()),
		),
	}
}

// getFieldPath returns the field path without the struct name prefix
// e.g. "TestConfig.Local.Options.Format" -> "Local.Options.Format"
// e.g. "Config[github.com/theplant/ciam-next/pkg/auth.testClaimsPayload].JWK" -> "JWK"
func getFieldPath(namespace string) string {
	// If the namespace contains a generic type, extract the field after "]."
	if strings.Contains(namespace, "[") && strings.Contains(namespace, "]") {
		if parts := strings.Split(namespace, "]."); len(parts) > 1 {
			return parts[1]
		}
		return namespace
	}

	parts := strings.Split(namespace, ".")
	if len(parts) > 1 {
		return strings.Join(parts[1:], ".")
	}
	return namespace
}

// RunTests runs a set of validation tests.
//
// It takes a slice of ExpectedValidation and runs each test case in a
// separate subtest. For each test case, it validates the given config using
// the embedded validator and checks that the validation errors match the
// expected errors. If the expected errors are empty, it checks that the config
// is valid.
//
// It also checks for duplicate expectations and unexpected errors.
func (s *ValidationSuite) RunTests(expectations []ExpectedValidation) {
	for i := range expectations {
		if expectations[i].Name == "" {
			s.t.Errorf("test case %d must have a name", i+1)
			return
		}
	}

	for _, exp := range expectations {
		// Check for duplicate expectations
		seen := make(map[string]map[string]bool)
		for _, exp := range exp.ExpectedErrors {
			if _, ok := seen[exp.Path]; !ok {
				seen[exp.Path] = make(map[string]bool)
			}
			if seen[exp.Path][exp.Tag] {
				panic(fmt.Sprintf("duplicate expectation found: path=%q tag=%q", exp.Path, exp.Tag))
			}
			seen[exp.Path][exp.Tag] = true
		}

		s.t.Run(exp.Name, func(t *testing.T) {
			err := s.validator.StructCtx(context.Background(), exp.Config)

			if len(exp.ExpectedErrors) == 0 {
				assert.NoError(t, err, "expected config to be valid")
				return
			}

			assert.Error(t, err, "expected config to be invalid")
			if err == nil {
				return
			}

			// Convert validation errors to a map for easier lookup
			var validationErrs validator.ValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Fatalf("expected validator.ValidationErrors, got %T", err)
			}

			errMap := make(map[string][]string)
			for _, verr := range validationErrs {
				path := getFieldPath(verr.Namespace())
				errMap[path] = append(errMap[path], verr.Tag())
			}

			// Check that all expected errors are present
			for _, expected := range exp.ExpectedErrors {
				tags, found := errMap[expected.Path]
				if !found {
					t.Errorf("expected validation error for path %q not found", expected.Path)
					continue
				}

				found = false
				for _, tag := range tags {
					if tag == expected.Tag {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("path %q: expected validation tag %q not found in %v",
						expected.Path, expected.Tag, tags)
				}
			}

			// Check for unexpected errors
			for path, tags := range errMap {
				expectedTags, pathExists := seen[path]
				if !pathExists {
					t.Errorf("unexpected validation error: path=%q", path)
					continue
				}
				for _, tag := range tags {
					if !expectedTags[tag] {
						t.Errorf("unexpected validation error: path=%q tag=%q", path, tag)
					}
				}
			}
		})
	}
}

// WithCustomValidator returns a new ValidationSuite that uses the provided validator.
// This is useful when you need to test custom validation rules.
func (s *ValidationSuite) WithCustomValidator(v Validator) *ValidationSuite {
	return &ValidationSuite{
		t:         s.t,
		validator: v,
	}
}
