package config

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/confx"
	"github.com/spf13/pflag"
)

// ServerConfig defines server-related configuration.
type ServerConfig struct {
	Host string `confx:"host" usage:"Server host address" validate:"required"`
	Port int    `confx:"port" usage:"Server port" validate:"gte=1,lte=65535"`
	TLS  bool   `confx:"tls" usage:"Enable TLS"`
}

// CommonDBConfig defines common database configuration settings.
type CommonDBConfig struct {
	Name     string        `confx:"name" usage:"Database name" validate:"required"`
	Username string        `confx:"username" usage:"Database username"`
	Password string        `confx:"password" usage:"Database password"`
	Timeout  time.Duration `confx:"timeout" usage:"Database connection timeout" validate:"gte=0"`
}

// DatabaseConfig defines database-related configuration.
type DatabaseConfig struct {
	Type string `confx:"type" usage:"Database type (postgres, sqlite)" validate:"required,oneof=postgres sqlite"`
	// Flatten the CommonDBConfig fields into this struct using squash
	CommonDBConfig `confx:",squash"`
	// Connection details that vary by database type
	Host string `confx:"host" usage:"Database host"`
	Port int    `confx:"port" usage:"Database port" validate:"omitempty,gte=1,lte=65535"`
}

// JWTConfig defines JWT authentication configuration.
type JWTConfig struct {
	Secret string `confx:"secret" usage:"JWT secret key" validate:"required"`
}

// OAuthConfig defines OAuth authentication configuration.
type OAuthConfig struct {
	ClientID     string `confx:"clientID" usage:"OAuth client ID" validate:"required"`
	ClientSecret string `confx:"clientSecret" usage:"OAuth client secret" validate:"required"`
}

// AuthConfig defines authentication configuration with skip_nested_unless validation.
type AuthConfig struct {
	Provider string      `confx:"provider" usage:"Authentication provider (jwt, oauth, basic)" validate:"required,oneof=jwt oauth basic"`
	JWT      JWTConfig   `confx:"jwt" validate:"skip_nested_unless=Provider jwt"`
	OAuth    OAuthConfig `confx:"oauth" validate:"skip_nested_unless=Provider oauth"`
}

// LoggingConfig defines logging configuration.
type LoggingConfig struct {
	Level  string `confx:"level" usage:"Log level (debug, info, warn, error)" validate:"required,oneof=debug info warn error"`
	Output string `confx:"output" usage:"Log output (stdout, file)" validate:"required,oneof=stdout file"`
	Path   string `confx:"path" usage:"Log file path" validate:"required_if=Output file"`
}

// Config is the main configuration structure.
type Config struct {
	Server         ServerConfig   `confx:"server" validate:"required"`
	Database       DatabaseConfig `confx:"database" validate:"required"`
	Auth           AuthConfig     `confx:"auth" validate:"required"`
	Logging        LoggingConfig  `confx:"logging" validate:"required"`
	aPrivatedField string         // ignored by confx
	PublicField    string         `confx:"-"` // ignored by confx
}

// maskSecret masks a secret string for display.
func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:2] + "****" + secret[len(secret)-2:]
}

// Print prints the configuration.
func (c *Config) Print() {
	fmt.Println("=== Server Configuration ===")
	fmt.Printf("Host: %s\n", c.Server.Host)
	fmt.Printf("Port: %d\n", c.Server.Port)
	fmt.Printf("TLS: %t\n", c.Server.TLS)

	fmt.Println("\n=== Database Configuration ===")
	fmt.Printf("Type: %s\n", c.Database.Type)
	if c.Database.Host != "" {
		fmt.Printf("Host: %s\n", c.Database.Host)
	}
	if c.Database.Port > 0 {
		fmt.Printf("Port: %d\n", c.Database.Port)
	}
	fmt.Printf("Database Name: %s\n", c.Database.Name)
	if c.Database.Username != "" {
		fmt.Printf("Username: %s\n", c.Database.Username)
	}
	if c.Database.Password != "" {
		fmt.Printf("Password: %s\n", maskSecret(c.Database.Password))
	}
	fmt.Printf("Timeout: %s\n", c.Database.Timeout)

	fmt.Println("\n=== Authentication Configuration ===")
	fmt.Printf("Provider: %s\n", c.Auth.Provider)
	if c.Auth.Provider == "jwt" {
		fmt.Printf("JWT Secret: %s\n", maskSecret(c.Auth.JWT.Secret))
	} else if c.Auth.Provider == "oauth" {
		fmt.Printf("OAuth Client ID: %s\n", c.Auth.OAuth.ClientID)
		fmt.Printf("OAuth Client Secret: %s\n", maskSecret(c.Auth.OAuth.ClientSecret))
	}

	fmt.Println("\n=== Logging Configuration ===")
	fmt.Printf("Level: %s\n", c.Logging.Level)
	fmt.Printf("Output: %s\n", c.Logging.Output)
	if c.Logging.Output == "file" {
		fmt.Printf("File Path: %s\n", c.Logging.Path)
	}
}

//go:embed embed/default.yaml
var defaultConfigYAML string

func Initialize(flagSet *pflag.FlagSet, envPrefix string) (confx.Loader[*Config], error) {
	def, err := confx.Read[*Config]("yaml", strings.NewReader(defaultConfigYAML))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default config from embedded YAML")
	}
	return confx.Initialize(def, confx.WithFlagSet(flagSet), confx.WithEnvPrefix(envPrefix))
}
