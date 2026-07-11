# AGENTS.md

This file provides guidance to AI coding assistants when working with code in this repository.

## Project Overview

ConfX is a Go configuration management library that unifies command-line flags, environment variables, configuration files (YAML/JSON/TOML), and default values. It uses reflection to automatically generate flags from struct fields and integrates with Viper and go-playground/validator.

## Recommended Usage Pattern (for consumers)

When a project depends on ConfX, the recommended way to supply defaults is **not** a hand-written default struct literal, but a YAML file embedded into the binary and loaded via `confx.Read`:

```go
//go:embed default-config.yaml
var defaultConfigYAML string

def, err := confx.Read[*Config]("yaml", strings.NewReader(defaultConfigYAML))
// ...
loader, err := confx.Initialize(def, opts...)
```

A committed, embedded default config file doubles as **living documentation** of the command's configuration:

- It lists every configuration option the command supports.
- It shows the default value of each option.
- Comments in the file describe what each option does.
- Users can copy it verbatim, then trim or override only what they need.

`examples/config` is the canonical implementation of this pattern; prefer it when generating new ConfX-based configuration code.

## Common Commands

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestName ./...

# Run examples
cd examples/basic && go run .
cd examples/cobra && go run . serve
cd examples/pflag && go run . --config=path/to/config.yaml

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

## Architecture

### Core Components

**confx.go**: The main entry point providing `Initialize[T](def T, options ...Option) (Loader[T], error)`. It uses reflection to traverse struct fields, automatically registers pflag flags, binds environment variables, and returns a loader function that parses and validates configuration.

**option.go**: Configuration options including `WithFlagSet`, `WithEnvPrefix`, `WithViper`, `WithValidator`, `WithTagName`, `WithUsageTagName`, and `WithFieldHook`.

**validator.go**: Wraps go-playground/validator with a custom `skip_nested_unless` validation tag. This allows conditional validation of nested structs based on parent field values (e.g., only validate JWT config when Provider="jwt").

**decode.go**: Utility functions `Read[T]` and `ReadWithTagName[T]` for loading configuration directly from files or readers without using the full initialization flow.

### Struct Tag Conventions

- `confx:"fieldName"` - Configuration key name (kebab-case flag, SCREAMING_SNAKE env var)
- `confx:",squash"` - Flatten nested struct fields into parent
- `confx:"-"` - Ignore this field
- `usage:"description"` - Flag usage description
- `validate:"..."` - go-playground/validator validation rules

### Key Design Patterns

1. **Pointer Auto-initialization**: Nil pointers in the default config are automatically initialized to zero values during loading since flags require non-nil defaults.

2. **sync.Once Parsing**: Flag parsing happens only once via sync.Once in the returned loader function.

3. **Validation Error Filtering**: The `skip_nested_unless` validator returns errors for skipped fields, which are then filtered out by `skipNestedUnlessWrapper` - this prevents validation of irrelevant nested configs.

## Code Standards

- All comments must be in English
- Use `github.com/pkg/errors` for error wrapping (wrap external errors, pass through internal ones)
- Use `any` instead of `interface{}`
- Acronyms in identifiers are uppercase (ID, HTTP, URL)
- JSON tags use camelCase
- String enum constants use UPPERCASE_WITH_UNDERSCORES for values
- Interface implementations must include validation declarations: `var _ InterfaceName = &ImplementationType{}`
- Include `context.Context` as first parameter in IO operations
