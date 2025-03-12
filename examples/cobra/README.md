# Cobra Example with Confx

This example demonstrates how to build a CLI application using [Cobra](https://github.com/spf13/cobra) with advanced configuration management powered by [Confx](https://github.com/qor5/confx).

## Overview

This example showcases:

- A CLI application with multiple commands (`serve` and `migrate`)
- Structured configuration with validation
- Loading configuration from multiple sources (files, environment variables, command-line flags)
- Type-safe configuration with Go structs

## Usage

### Building the Application

```bash
go build -o demo .
```

### Available Commands

#### Root Command

```bash
./demo
```

Shows the available commands and general help information.

#### Serve Command

```bash
# Show help
./demo serve -h

# Start server with default configuration
./demo serve

# Start server with a specific configuration file
./demo serve --config ../config/sample.yaml
```

#### Migrate Command

```bash
# Run migration with default database settings
./demo migrate

# Run migration with a specific DSN
./demo migrate --database-dsn "postgres://user:pass@localhost:5432/dbname"
```

### Configuration

Configuration can be provided in multiple ways:

1. **Default values**: Embedded in the application
2. **Configuration file**: Specified with `--config` flag
3. **Environment variables**: Prefixed with `DEMO_`
4. **Command-line flags**: Take precedence over other methods

Example environment variables:

```bash
DEMO_SERVER_HOST=localhost
DEMO_SERVER_PORT=8080
DEMO_DATABASE_USERNAME=admin
```

## Configuration Structure

The configuration is type-safe and defined as Go structs with validation:

- **Server**: Host, port, and TLS settings
- **Database**: Type, credentials, and connection settings
- **Auth**: Provider (JWT, OAuth, Basic) with appropriate settings
- **Logging**: Level, output location, and file path

See `../config/config.go` for detailed field definitions and validation rules.

## Sample Configuration

A sample configuration is provided in `../config/sample.yaml`. You can use this as a template
for your own configuration, or modify it to match your needs:

```yaml
server:
  host: "0.0.0.0"
  port: 9090
  tls: true

database:
  type: "sqlite"
  name: "/var/lib/myapp/data.db"

auth:
  provider: "oauth"
  oauth:
    clientID: "example-client-id"
    clientSecret: "example-client-secret"

logging:
  level: "warn"
  output: "file"
  path: "/var/log/myapp.log"
```

## License

This example is part of the QOR5 Confx project, see the main repository for license information.
