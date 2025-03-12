# Basic Example

This example demonstrates the core features of the `confx` library for configuration management in Go applications.

## Key Features Demonstrated

This example showcases the following `confx` features:

1. **Multi-source Configuration**: Load configurations from default values, environment variables, command-line arguments, and YAML files.

2. **Default Configuration**: Load default configuration from an embedded YAML string.

3. **Embedded Configuration**: Use the `squash` tag to flatten embedded structs for configuration reuse.

4. **Conditional Validation**: Validate configuration using struct tags with conditions:

   - `required_if` for simple conditional validation
   - `skip_nested_unless` for conditional nested struct validation

5. **Configuration Testing**: Test configuration validation logic using the validation suite.

## Configuration Structure

The configuration is organized into the following sections:

- **Server**: Basic server settings (host, port, TLS)
- **Database**: Database connection settings, using `squash` tag to embed CommonDBConfig
- **Auth**: Authentication settings with nested configurations:
  - JWT configuration (validated only when provider is "jwt")
  - OAuth configuration (validated only when provider is "oauth")
- **Logging**: Logging configuration

## Running the Example

1. Show help

   ```bash
     go run main.go -h
   ```

2. Run with default configuration:

   ```bash
   go run main.go
   ```

3. Run with custom configuration file:

   ```bash
   go run main.go --config ../config/sample.yaml
   ```

4. Override specific values via command line:

   ```bash
   go run main.go --server-port 9090 --auth-provider oauth --auth-oauth-client-id "my-client" --auth-oauth-client-secret "my-secret"
   ```

5. Override specific values via environment variables:

   ```bash
   APP_SERVER_PORT=9090 APP_AUTH_PROVIDER=oauth APP_AUTH_OAUTH_CLIENT_ID="ev-client" APP_AUTH_OAUTH_CLIENT_SECRET="ev-secret" go run main.go
   ```

## Conditional Validation Examples

This example demonstrates two types of conditional validation:

### 1. Simple Field Conditional Validation

- Log file path is only required when the output is set to "file" (`required_if=Output file`)

### 2. Nested Struct Conditional Validation

Using `skip_nested_unless` for conditional validation of nested structures:

- JWT configuration is only validated when the auth provider is "jwt"
- OAuth configuration is only validated when the auth provider is "oauth"

This is particularly useful for complex configurations where entire sections should be conditionally validated based on parent field values.
