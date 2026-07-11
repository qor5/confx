# AGENTS.md

This file provides guidance to AI coding assistants when working with code in this repository. It is the single source of truth for both consumers of the library and maintainers of it.

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

`Initialize[T]` takes a default config struct plus options and returns a `Loader[T]` function. Initialization recursively walks the struct to register pflag flags and collect viper bindings; the returned loader (on first call) parses flags, binds env vars, reads the optional config file, unmarshals into the struct, and validates it.

```
┌─────────────────────────────────────────────────────────────┐
│                        Initialize[T]                         │
│  Input: default config + options → Output: Loader[T] func    │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Option    │ │  Recursive  │ │   Loader    │
│   Config    │ │ Initialize  │ │   Function  │
└─────────────┘ └──────┬──────┘ └──────┬──────┘
                       │               │
              ┌────────┴────────┐      │
              ▼                 ▼      │
        ┌──────────┐      ┌──────────┐ │
        │  pflag   │      │  viper   │ │
        │ Register │      │  Bind    │ │
        └──────────┘      └──────────┘ │
                                       │
                              ┌────────┴────────┐
                              ▼                 ▼
                        ┌──────────┐      ┌──────────┐
                        │  Config  │      │ Validator│
                        │   File   │      │          │
                        └──────────┘      └──────────┘
```

### Core File Responsibilities

| File | Responsibility | Key Functions/Types |
|------|----------------|---------------------|
| `confx.go` | Core implementation: reflection walk, flag registration, deferred binding, the loader | `Initialize`, `initializeRecursive`, `unwrapOrNew`, `Loader` |
| `option.go` | Configuration options | `Option`, `initOptions`, `WithFlagSet`, `WithEnvPrefix`, `WithViper`, `WithValidator`, `WithTagName`, `WithUsageTagName`, `WithFieldHook` |
| `decode.go` | Standalone decoding + mapstructure hooks | `Read`, `ReadWithTagName`, `StringToSliceHookFunc`, `StringToMapHookFunc` |
| `validator.go` | Conditional-validation extension | `ValidatorWithSkipNestedUnless`, `skipNestedUnlessImpl`, `skipNestedUnlessWrapper` |
| `validation_suite.go` | Test helper for validation | `ValidationSuite`, `NewValidationSuite`, `RunTests`, `ExpectedValidation` |

### Struct Tag Conventions

- `confx:"fieldName"` — Configuration key name (kebab-case flag, SCREAMING_SNAKE env var)
- `confx:",squash"` — Flatten nested struct fields into parent (non-pointer structs only)
- `confx:"-"` — Ignore this field (also applies implicitly to unexported fields)
- `usage:"description"` — Flag usage description
- `validate:"..."` — go-playground/validator validation rules

### Naming Conversion Rules

```go
// Struct path: Database.Host
viperKey = "database.host"       // Viper internal key
flagKey  = "database-host"       // CLI flag --database-host
envKey   = "APP_DATABASE_HOST"   // Environment variable (with prefix "APP_")
```

- `viperKey`: join segments with `.`, keeping the original tag name.
- `flagKey`: convert with `lo.KebabCase`, fixing a leading hyphen before numbers.
- `envKey`: replace `.` and `-` with `_`, uppercase, then prepend the env prefix.

### Type Support Matrix

| Type | pflag Register | Env | Config File | Special Handling |
|------|----------------|-----|-------------|------------------|
| `bool` | `Bool` | ✓ | ✓ | - |
| `int*` | `Int/Int8/16/32/64` | ✓ | ✓ | - |
| `uint*` | `Uint/Uint8/16/32/64` | ✓ | ✓ | - |
| `float*` | `Float32/64` | ✓ | ✓ | - |
| `string` | `String` | ✓ | ✓ | - |
| `time.Duration` | `Duration` | ✓ | ✓ | Recognized as a distinct type |
| `time.Time` | `String` (RFC3339) | ✓ | ✓ | Requires format conversion |
| `[]T` | `XxxSlice` | ✓ | ✓ | Struct slices are encoded as JSON |
| `[]byte` | `BytesBase64` | ✓ | ✓ | Base64 encoding |
| `map[string]T` | `StringToXxx` | ✓ | ✓ | Only int/int64/string values |
| `struct` | Recursive | ✓ | ✓ | Nested processing |

## Key Design Patterns

1. **Recursive initialization (`initializeRecursive`)**: iterate struct fields — skip unexported fields and `confx:"-"`; recurse into `,squash` (flattened) and nested structs (with a parent key); for basic types, register a pflag and append a bind function to `collectBinds`.

2. **Deferred binding via `sync.Once`**: bindings are collected as `[]func() error` during `Initialize` but executed only on the loader's first call, once. This is required because `BindPFlag` needs the flagSet parsed first, binding must happen exactly once, and it must support a custom flagSet (e.g. Cobra).

3. **Automatic pointer initialization (`unwrapOrNew`)**: nil pointers in the default config are dereferenced to fresh zero values during the walk, so every field is a usable non-pointer value — flags require non-nil defaults.

4. **`skip_nested_unless` conditional validation**: a custom tag that conditionally validates an entire nested struct. During validation, when the condition doesn't match, `skipNestedUnlessImpl` returns false (producing an error), and `skipNestedUnlessWrapper` then filters out those skip-related errors — so an irrelevant nested config is effectively not validated.

   ```go
   Provider string    `confx:"provider" validate:"required,oneof=jwt oauth"`
   JWT      JWTConfig `confx:"jwt" validate:"skip_nested_unless=Provider jwt"`
   // JWT is only validated when Provider == "jwt"
   ```

## Modifying the Code

### Adding a New Option

1. Add a field to the `initOptions` struct in `option.go`.
2. Implement the `WithXxx` function (check for nil/empty).
3. Apply the option where `initOptions` is consumed in `confx.go`.

### Adding New Type Support

1. Add a case in the type switch of `initializeRecursive` (`confx.go`).
2. Implement the corresponding `flagSet.Xxx` registration.
3. Add a decode hook in `decode.go` if the type needs custom string decoding.
4. Update the Type Support Matrix above.

### Modifying Validation Logic

- Basic validation: standard go-playground/validator tags.
- Conditional validation: use the existing `skip_nested_unless`.
- New custom tags: follow the pattern in `validator.go`.

### Testing Requirements

- Unit tests covering both normal and error paths.
- Integration tests using the `examples/config` configuration.
- New options: test option combinations.

Use `ValidationSuite` for validation tests (see `validation_suite_test.go` for full examples):

```go
suite := NewValidationSuite(t)
suite.RunTests([]ExpectedValidation{
    {Name: "valid config", Config: &Config{ /* ... */ }}, // no ExpectedErrors → expected to pass
    {Name: "invalid port", Config: &Config{ /* ... */ }, ExpectedErrors: []ExpectedValidationError{ /* ... */ }},
})
```

## Common Pitfalls

1. **flagSet parsing timing**: the default flagSet is auto-parsed inside the loader; a custom flagSet must be parsed by the caller before invoking the loader (Cobra parses it for you).
2. **Viper key conflicts**: Viper uses `.` as a separator — tag names must not contain `.`.
3. **Slice defaults**: pflag slice values given on the CLI *replace* the default entirely, they are not appended.
4. **Squash + pointers**: `,squash` only works on non-pointer structs; nil pointers elsewhere are auto-initialized by `unwrapOrNew`.
5. **Env prefix**: the prefix must end with an underscore (e.g. `APP_`), otherwise `APP_DATABASE_HOST` becomes `APPDATABASE_HOST`.

## Debugging Tips

```go
// After Initialize, inspect registered flags:
opts.flagSet.VisitAll(func(f *pflag.Flag) {
    fmt.Printf("Flag: --%s (default: %v)\n", f.Name, f.DefValue)
})

// Inspect resolved viper state:
viperInstance.AllSettings()
viperInstance.GetString("database.host")
```

## Code Standards

- All comments must be in English.
- Use `github.com/pkg/errors` for error wrapping (wrap external errors, pass through internal ones).
- Use `any` instead of `interface{}`.
- Acronyms in identifiers are uppercase (ID, HTTP, URL).
- JSON tags use camelCase.
- String enum constants use UPPERCASE_WITH_UNDERSCORES for values.
- Interface implementations must include a compile-time assertion: `var _ InterfaceName = &ImplementationType{}`.
- Include `context.Context` as the first parameter in IO operations.

## Dependencies

Core (imported by the library itself):

```
spf13/viper                    Configuration management core
spf13/pflag                    Command-line flags
go-playground/validator/v10    Validation
go-viper/mapstructure/v2       Struct decoding
spf13/cast                     Type conversion in decode hooks
huandu/go-clone                Deep clone of the default config
samber/lo                      Utility functions
pkg/errors                     Error wrapping
```

`spf13/cobra` and `stretchr/testify` are used only by the examples and tests.

## Release Process

1. Run all tests: `go test ./...`.
2. Verify the examples still run.
3. Update the CHANGELOG if one exists.
4. Tag the release: `git tag vX.Y.Z && git push origin vX.Y.Z`.
