// Package confx unifies command-line flags, environment variables, configuration
// files (YAML/JSON/TOML), and struct-defined defaults into a single typed loader.
//
// # Recommended: embed a default config file as living documentation
//
// The recommended way to supply defaults is to keep them in a YAML file that you
// embed into the binary and load via [Read], rather than hand-writing a default
// struct literal:
//
//	//go:embed default-config.yaml
//	var defaultConfigYAML string
//
//	def, err := confx.Read[*Config]("yaml", strings.NewReader(defaultConfigYAML))
//	if err != nil {
//		return err
//	}
//	loader, err := confx.Initialize(def)
//
// A committed, embedded default config file doubles as living documentation of the
// command's configuration:
//
//   - It lists every configuration option the command supports.
//   - It shows the default value of each option.
//   - Comments in the file describe what each option does.
//   - Users can copy it verbatim, then trim or override only what they need.
//
// See examples/config for a complete, working example of this pattern.
package confx
