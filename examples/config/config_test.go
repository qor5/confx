package config

import (
	"strings"
	"testing"
	"time"

	"github.com/qor5/confx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigValidation tests validation of configuration using validation_suite.
func TestConfigValidation(t *testing.T) {
	// Create a validation test suite
	suite := confx.NewValidationSuite(t)

	// Define test expectations
	expectations := []confx.ExpectedValidation{
		{
			Name:           "Valid default configuration",
			Config:         getValidConfig(),
			ExpectedErrors: []confx.ExpectedValidationError{
				// Default config should pass validation
			},
		},
		{
			Name:   "Missing required server host",
			Config: getConfigWithMissingHost(),
			ExpectedErrors: []confx.ExpectedValidationError{
				{Path: "Server.Host", Tag: "required"},
			},
		},
		{
			Name:   "Invalid port range",
			Config: getConfigWithInvalidPort(),
			ExpectedErrors: []confx.ExpectedValidationError{
				{Path: "Server.Port", Tag: "lte"},
			},
		},
		{
			Name:   "JWT validation when provider is jwt",
			Config: getConfigWithInvalidJWT(),
			ExpectedErrors: []confx.ExpectedValidationError{
				{Path: "Auth.JWT.Secret", Tag: "required"},
			},
		},
		{
			Name:   "OAuth validation when provider is oauth",
			Config: getConfigWithInvalidOAuth(),
			ExpectedErrors: []confx.ExpectedValidationError{
				{Path: "Auth.OAuth.ClientID", Tag: "required"},
			},
		},
		{
			Name:           "JWT ignored when provider is not jwt",
			Config:         getConfigWithBasicAuthAndInvalidJWT(),
			ExpectedErrors: []confx.ExpectedValidationError{
				// No JWT errors due to skip_nested_unless
			},
		},
		{
			Name:   "File path required when output is file",
			Config: getConfigWithFileOutputNoPath(),
			ExpectedErrors: []confx.ExpectedValidationError{
				{Path: "Logging.Path", Tag: "required_if"},
			},
		},
		{
			Name:           "SQLite database without host/port is valid",
			Config:         getConfigWithSQLiteDatabase(),
			ExpectedErrors: []confx.ExpectedValidationError{
				// No errors expected
			},
		},
	}

	// Run the tests
	suite.RunTests(expectations)
}

// TestReadFromYAML tests loading config from YAML.
func TestReadFromYAML(t *testing.T) {
	yamlConfig := `
server:
  host: "test.host"
  port: 9090
`
	config, err := confx.Read[*Config]("yaml", strings.NewReader(yamlConfig))
	require.NoError(t, err)
	assert.Equal(t, "test.host", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
}

// getValidConfig returns a valid configuration for testing.
func getValidConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			TLS:  false,
		},
		Database: DatabaseConfig{
			Type: "postgres",
			CommonDBConfig: CommonDBConfig{
				Name:     "myapp",
				Username: "postgres",
				Password: "postgres",
				Timeout:  10 * time.Second,
			},
			Host: "localhost",
			Port: 5432,
		},
		Auth: AuthConfig{
			Provider: "jwt",
			JWT: JWTConfig{
				Secret: "test-secret",
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "stdout",
			Path:   "",
		},
	}
}

// getConfigWithMissingHost returns a config with missing server host.
func getConfigWithMissingHost() Config {
	config := getValidConfig()
	config.Server.Host = ""
	return config
}

// getConfigWithInvalidPort returns a config with invalid port.
func getConfigWithInvalidPort() Config {
	config := getValidConfig()
	config.Server.Port = 70000
	return config
}

// getConfigWithInvalidJWT returns a config with JWT provider but missing secret.
func getConfigWithInvalidJWT() Config {
	config := getValidConfig()
	config.Auth.Provider = "jwt"
	config.Auth.JWT.Secret = ""
	return config
}

// getConfigWithInvalidOAuth returns a config with OAuth provider but missing clientID.
func getConfigWithInvalidOAuth() Config {
	config := getValidConfig()
	config.Auth.Provider = "oauth"
	config.Auth.OAuth = OAuthConfig{
		ClientSecret: "secret",
	}
	return config
}

// getConfigWithBasicAuthAndInvalidJWT returns a config with basic auth and invalid JWT.
func getConfigWithBasicAuthAndInvalidJWT() Config {
	config := getValidConfig()
	config.Auth.Provider = "basic"
	config.Auth.JWT.Secret = ""
	return config
}

// getConfigWithFileOutputNoPath returns a config with file output but no path.
func getConfigWithFileOutputNoPath() Config {
	config := getValidConfig()
	config.Logging.Output = "file"
	config.Logging.Path = ""
	return config
}

// getConfigWithSQLiteDatabase returns a config with SQLite database.
func getConfigWithSQLiteDatabase() Config {
	config := getValidConfig()
	config.Database.Type = "sqlite"
	config.Database.Host = "" // Not required for SQLite
	config.Database.Port = 0  // Not required for SQLite
	config.Database.Name = "app.db"
	return config
}
