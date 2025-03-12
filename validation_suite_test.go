package confx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name   string     `validate:"required"`
	Age    int        `validate:"gte=0,lte=150"`
	Type   string     `validate:"oneof=local remote"`
	Local  LocalConf  `validate:"skip_nested_unless=Type local"`
	Remote RemoteConf `validate:"skip_nested_unless=Type remote"`
}

type LocalConf struct {
	Path    string `validate:"required"`
	Options struct {
		Cache  bool   `validate:"required"`
		Format string `validate:"oneof=json yaml"`
	} `validate:"required"`
}

type RemoteConf struct {
	URL      string `validate:"required,url"`
	Username string `validate:"required_with=Password"`
	Password string `validate:"required_with=Username"`
	Auth     struct {
		Type   string `validate:"required,oneof=basic token"`
		Token  string `validate:"required_if=Type token"`
		APIKey string `validate:"required_if=Type basic"`
	} `validate:"required"`
}

func TestValidationSuite(t *testing.T) {
	suite := NewValidationSuite(t)

	tests := []ExpectedValidation{
		{
			Name: "valid config/type=local",
			Config: &TestConfig{
				Name: "John",
				Age:  30,
				Type: "local",
				Local: LocalConf{
					Path: "/tmp/data",
					Options: struct {
						Cache  bool   `validate:"required"`
						Format string `validate:"oneof=json yaml"`
					}{
						Cache:  true,
						Format: "json",
					},
				},
			},
		},
		{
			Name: "valid config/type=remote",
			Config: &TestConfig{
				Name: "Alice",
				Age:  25,
				Type: "remote",
				Remote: RemoteConf{
					URL:      "https://example.com",
					Username: "alice",
					Password: "secret",
					Auth: struct {
						Type   string `validate:"required,oneof=basic token"`
						Token  string `validate:"required_if=Type token"`
						APIKey string `validate:"required_if=Type basic"`
					}{
						Type:   "basic",
						APIKey: "key123",
					},
				},
			},
		},
		{
			Name: "invalid config/type=local",
			Config: &TestConfig{
				Age:  -1,
				Type: "local",
				Local: LocalConf{
					Options: struct {
						Cache  bool   `validate:"required"`
						Format string `validate:"oneof=json yaml"`
					}{
						Format: "xml",
					},
				},
			},
			ExpectedErrors: []ExpectedValidationError{
				{Path: "Name", Tag: "required"},
				{Path: "Age", Tag: "gte"},
				{Path: "Local.Path", Tag: "required"},
				{Path: "Local.Options.Cache", Tag: "required"},
				{Path: "Local.Options.Format", Tag: "oneof"},
			},
		},
		{
			Name: "invalid config/type=remote",
			Config: &TestConfig{
				Name: "Bob",
				Age:  30,
				Type: "remote",
				Remote: RemoteConf{
					URL: "not-a-url",
					Auth: struct {
						Type   string `validate:"required,oneof=basic token"`
						Token  string `validate:"required_if=Type token"`
						APIKey string `validate:"required_if=Type basic"`
					}{
						Type: "token",
					},
				},
			},
			ExpectedErrors: []ExpectedValidationError{
				{Path: "Remote.URL", Tag: "url"},
				{Path: "Remote.Auth.Token", Tag: "required_if"},
			},
		},
	}

	suite.RunTests(tests)
}

func TestExpectationWithoutName(t *testing.T) {
	mockT := new(testing.T)
	suite := NewValidationSuite(mockT)

	suite.RunTests([]ExpectedValidation{
		{
			Config: &TestConfig{
				Name: "John",
			},
			Name: "", // should fail
		},
	})

	require.True(t, mockT.Failed(), "Expectation without name should fail")
}

func TestMultipleTagValidation(t *testing.T) {
	suite := NewValidationSuite(t)

	type MultiTagConfig struct {
		// Test required + min/max length
		Name string `validate:"required,min=3,max=10"`
		// Test required + numeric range
		Age int `validate:"required,gte=18,lte=100"`
		// Test required + oneof + contains
		Type string `validate:"required,oneof=admin user guest,contains=user"`
		// Test multiple independent validations
		Email string `validate:"required,email,contains=@company.com"`
		// Test nested struct with multiple validations
		Settings struct {
			Theme    string `validate:"required,oneof=light dark"`
			Language string `validate:"required,oneof=en zh ja"`
			Display  struct {
				FontSize int    `validate:"required,gte=8,lte=32"`
				Color    string `validate:"required,hexcolor"`
			} `validate:"required"`
		} `validate:"required"`
	}

	tests := []ExpectedValidation{
		{
			Name: "invalid config with multiple tags",
			Config: &MultiTagConfig{
				Name:  "a",
				Age:   15,
				Type:  "superadmin",
				Email: "invalid",
				Settings: struct {
					Theme    string `validate:"required,oneof=light dark"`
					Language string `validate:"required,oneof=en zh ja"`
					Display  struct {
						FontSize int    `validate:"required,gte=8,lte=32"`
						Color    string `validate:"required,hexcolor"`
					} `validate:"required"`
				}{
					Theme:    "blue",
					Language: "invalid",
					Display: struct {
						FontSize int    `validate:"required,gte=8,lte=32"`
						Color    string `validate:"required,hexcolor"`
					}{
						FontSize: 6,
						Color:    "not-a-color",
					},
				},
			},
			ExpectedErrors: []ExpectedValidationError{
				{Path: "Name", Tag: "min"},
				{Path: "Age", Tag: "gte"},
				{Path: "Type", Tag: "oneof"},
				{Path: "Email", Tag: "email"},
				{Path: "Settings.Theme", Tag: "oneof"},
				{Path: "Settings.Language", Tag: "oneof"},
				{Path: "Settings.Display.FontSize", Tag: "gte"},
				{Path: "Settings.Display.Color", Tag: "hexcolor"},
			},
		},
		{
			Name: "valid config with multiple tags",
			Config: &MultiTagConfig{
				Name:  "Bob",
				Age:   20,
				Type:  "user",
				Email: "bob@company.com",
				Settings: struct {
					Theme    string `validate:"required,oneof=light dark"`
					Language string `validate:"required,oneof=en zh ja"`
					Display  struct {
						FontSize int    `validate:"required,gte=8,lte=32"`
						Color    string `validate:"required,hexcolor"`
					} `validate:"required"`
				}{
					Theme:    "light",
					Language: "en",
					Display: struct {
						FontSize int    `validate:"required,gte=8,lte=32"`
						Color    string `validate:"required,hexcolor"`
					}{
						FontSize: 16,
						Color:    "#FF0000",
					},
				},
			},
		},
	}

	suite.RunTests(tests)
}

func TestDuplicateExpectations(t *testing.T) {
	suite := NewValidationSuite(t)

	type Config struct {
		Field string `validate:"required,min=3"`
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on duplicate expectations, but got none")
		}
		if msg, ok := r.(string); ok {
			if !strings.Contains(msg, "duplicate expectation found") {
				t.Errorf("expected panic message to contain 'duplicate expectation found', but got: %s", msg)
			}
		}
	}()

	tests := []ExpectedValidation{
		{
			Name: "multiple expectations with same config",
			Config: &Config{
				Field: "a",
			},
			ExpectedErrors: []ExpectedValidationError{
				{Path: "Field", Tag: "min"},
				{Path: "Field", Tag: "min"}, // Duplicate expectation
			},
		},
	}

	suite.RunTests(tests)
}
