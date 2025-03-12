# confx - Configuration Management Library for Go

confx is a feature-rich configuration management library for Go that provides a comprehensive solution for handling application configuration. It combines command line arguments, environment variables, configuration files, and default values to help developers efficiently manage application configuration.

## Features

- **Unified Configuration Management**: Automatically binds command line flags, environment variables, and configuration files
- **Strong Type Support**: Use structs to define configuration with type-safe access
- **Rich Data Types**: Support for basic types, slices, maps, nested structs, and more
- **Pointer Type Support**: Auto-handles nil pointers to ensure all fields have usable values after configuration loading
- **Tag-Driven**: Define configuration key names, usage descriptions, and more through struct tags
- **Complete Validation Support**: Integrates with go-playground/validator, supporting all its validation rules and features
- **Enhanced Conditional Validation**: Extends standard validator with enhanced nested struct conditional validation
- **Struct Embedding**: Flatten nested struct fields using the `squash` tag
- **Customizable Options**: Flexible options to customize configuration loading behavior
- **Universal Configuration Reading**: Support for loading configuration from various file formats

## Installation

```bash
go get github.com/qor5/confx
```

## Quick Start

### Create a Default Configuration File

First, create a YAML file with your default configuration, named `default-config.yaml`:

```yaml
# default-config.yaml
server:
  host: localhost
  port: 8080
  timeout: 30s

database:
  host: localhost
  port: 5432
  username: user
  password: password
  database: myapp

logLevel: info
```

### Define Configuration Structs

```go
package main

import (
    "context"
    "embed"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/qor5/confx"
)

//go:embed default-config.yaml
var defaultConfigYaml string

// Define configuration structs
type ServerConfig struct {
    Host    string `confx:"host" usage:"Server host address" validate:"required"`
    Port    int    `confx:"port" usage:"Server port" validate:"gte=1,lte=65535"`
    Timeout time.Duration `confx:"timeout" usage:"Request timeout duration" validate:"gte=0"`
}

type DatabaseConfig struct {
    Host     string `confx:"host" usage:"Database host address" validate:"required"`
    Port     int    `confx:"port" usage:"Database port" validate:"gte=1,lte=65535"`
    Username string `confx:"username" usage:"Database username"`
    Password string `confx:"password" usage:"Database password"`
    Database string `confx:"database" usage:"Database name" validate:"required"`
}

type Config struct {
    Server   ServerConfig   `confx:"server" validate:"required"`
    Database DatabaseConfig `confx:"database" validate:"required"`
    LogLevel string         `confx:"logLevel" usage:"Logging level" validate:"oneof=debug info warn error"`
}

func main() {
    // Read default configuration from embedded YAML string
    // We typically embed default config in the binary for three benefits:
    // 1. CLI can run independently without external config files
    // 2. The file can be delivered to users to understand available config options
    // 3. Users can copy and modify the file, simplifying custom configuration

    // Note: If your configuration is simple enough, you can directly construct Config object
    // Example: defaultConfig := Config{Server: ServerConfig{Host: "localhost", Port: 8080}, ...}
    defaultConfig, err := confx.Read[Config]("yaml", strings.NewReader(defaultConfigYaml))
    if err != nil {
        log.Fatalf("Failed to read default config: %v", err)
    }

    // Initialize config loader
    loader, err := confx.Initialize(defaultConfig)
    if err != nil {
        log.Fatalf("Failed to initialize config loader: %v", err)
    }

    // Load configuration
    config, err := loader(context.Background(), "")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Use the configuration
    fmt.Printf("Server config: %s:%d\n", config.Server.Host, config.Server.Port)
    fmt.Printf("Database config: %s:%d/%s\n", config.Database.Host, config.Database.Port, config.Database.Database)
    fmt.Printf("Log level: %s\n", config.LogLevel)
}
```

### Command Line Flags

confx automatically generates command line flags for each field in your configuration struct:

```bash
./myapp --server-host=127.0.0.1 --server-port=9090 --log-level=debug
```

You can also specify a configuration file using the custom flag we added:

```bash
./myapp --config=production.yaml
```

Using the `WithFlagSet` option allows you to provide a custom FlagSet, which is particularly useful in the following scenarios:

1. When you need to integrate with an existing command-line argument system
2. When you want to customize the sorting or grouping of flags
3. When you need to add additional command-line parameters not related to configuration
4. When using subcommands in complex applications (such as when used with Cobra)

For example, you can create a custom FlagSet and add extra flags:

```go
flagSet := pflag.NewFlagSet("myapp", pflag.ContinueOnError)
flagSet.SortFlags = false

// Add custom flags
flagSet.StringVar(&configPath, "custom-config-flag", "", "Path to configuration file")
flagSet.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

// Initialize config loader with custom FlagSet
loader, err := confx.Initialize(defaultConfig, confx.WithFlagSet(flagSet))

// Load configuration
config, err := loader(context.Background(), configPath)
```

This allows you to have complete control over how command-line arguments are handled while still leveraging confx's automatic binding functionality.

### Environment Variables

confx also binds environment variables to configuration fields:

```bash
SERVER_HOST=127.0.0.1 SERVER_PORT=9090 LOG_LEVEL=debug ./myapp
```

You can customize the environment variable prefix using the `WithEnvPrefix` option:

```go
loader, err := confx.Initialize(defaultConfig, confx.WithEnvPrefix("APP_"))
```

Then use environment variables with the prefix:

```bash
APP_SERVER_HOST=127.0.0.1 APP_SERVER_PORT=9090 APP_LOG_LEVEL=debug ./myapp
```

### Configuration Files

confx supports loading configuration from various file formats:

```yaml
# config.yaml
server:
  host: 127.0.0.1
  port: 9090
  timeout: 60s

database:
  host: db.example.com
  port: 5432
  username: admin
  password: secret
  database: production

logLevel: debug
```

## Advanced Features

### Validation Features

confx fully integrates go-playground/validator, supporting all its built-in validation rules and features. Additionally, confx provides enhanced functionality.

#### Standard Validator Features

Here are examples of common validation features provided by go-playground/validator:

```go
type Config struct {
    // Basic validation rules
    Port      int       `validate:"required,gte=1,lte=65535"`
    Email     string    `validate:"required,email"`
    URL       string    `validate:"url"`
    CreatedAt time.Time `validate:"required"`

    // Conditional validation
    OutputPath string `validate:"required_if=OutputType file"` // Required only when OutputType is "file"

    // Slice validation
    Tags []string `validate:"required,min=1,dive,required"`
}
```

#### confx Enhanced Conditional Validation

confx extends the standard validator with the `skip_nested_unless` validation rule for conditionally validating entire nested structures:

```go
type AuthConfig struct {
    Provider string    `confx:"provider" validate:"required,oneof=jwt oauth basic"`
    // Only validate JWT config when Provider is "jwt" - This is a confx enhancement
    JWT      JWTConfig `confx:"jwt" validate:"skip_nested_unless=Provider jwt"`
    // Only validate OAuth config when Provider is "oauth" - This is a confx enhancement
    OAuth    OAuthConfig `confx:"oauth" validate:"skip_nested_unless=Provider oauth"`
}
```

### Struct Embedding

You can use the `squash` tag to flatten nested struct fields into the parent struct:

```go
type CommonDBConfig struct {
    Name     string `confx:"name" validate:"required"`
    Username string `confx:"username"`
    Password string `confx:"password"`
}

type DatabaseConfig struct {
    Type string `confx:"type" validate:"required,oneof=postgres sqlite"`
    // Flatten CommonDBConfig fields into this struct
    CommonDBConfig `confx:",squash"`
    // Database-specific fields
    Host string `confx:"host"`
    Port int    `confx:"port" validate:"omitempty,gte=1,lte=65535"`
}
```

### Ignoring Fields

Use the `confx:"-"` tag to have confx completely ignore certain fields in your struct. These fields won't be mapped, won't generate flags, and won't be overridden by environment variables:

```go
type Config struct {
    // Normal field that confx will process
    Database DatabaseConfig `confx:"database"`

    // Ignored field - won't be processed by confx
    InternalState string `confx:"-"`

    // Private fields are automatically ignored (no explicit tag needed)
    internalCache map[string]interface{}

    // Even exported fields can be ignored with the "-" tag
    HelperFunction func() `confx:"-"`
}
```

### Custom Options

confx provides various options to customize configuration loading behavior:

```go
loader, err := confx.Initialize(defaultConfig,
    confx.WithEnvPrefix("APP_"),           // Set environment variable prefix
    confx.WithFlagSet(customFlagSet),      // Use custom FlagSet
    confx.WithViper(customViper),          // Use custom Viper instance
    confx.WithValidator(customValidator),  // Use custom validator
    confx.WithTagName("custom"),           // Use custom struct tag name
    confx.WithUsageTagName("description"), // Use custom usage tag name
    confx.WithFieldHook(customFieldHook),  // Custom field processing
)
```

## Utility Functions

### Direct Configuration Loading

In addition to the Initialize method, confx provides simple functions to load configuration directly from files or readers:

```go
// Load from config file
config, err := confx.Read[Config]("yaml", configFile)

// Load using custom tag name
config, err := confx.ReadWithTagName[Config]("custom", "yaml", configFile)
```

## Integration with Viper and Cobra

confx seamlessly integrates with the popular Viper and Cobra libraries:

- **Viper**: confx uses Viper as the underlying configuration management engine and allows you to use a custom Viper instance via the `WithViper` option
- **Cobra**: Check the `examples/cobra` directory to learn how to integrate confx with the Cobra command line framework

## Examples

Check the `examples` directory for more examples:

- `examples/basic`: Basic usage example
- `examples/cobra`: Integration with Cobra
- `examples/config`: Shared configuration package used by examples

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests to this repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
