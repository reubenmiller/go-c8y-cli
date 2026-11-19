// Command gen-surface walks the live c8y command tree and emits the
// projection-neutral CLI surface as JSON (commands, flags, and the metadata
// each carries). It is the clean seam described in
// proposals/CLI_CODEGEN_INVERSION.md: a stable manifest that downstream
// projectors (PowerShell, tests, docs) and CI diffs can consume without re-deriving
// the tree walk.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clibuild"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/clisurface"
)

func main() {
	command := flag.String("command", "", "Limit the surface to this command subtree (e.g. \"devices\")")
	out := flag.String("out", "", "Write the surface JSON to this file instead of stdout")
	flag.Parse()

	if err := run(*command, *out); err != nil {
		fmt.Fprintf(os.Stderr, "gen-surface: %s\n", err)
		os.Exit(1)
	}
}

func run(command, out string) error {
	root, err := clibuild.NewRootCommand()
	if err != nil {
		return err
	}
	root.InitDefaultHelpCmd()

	commands := clisurface.Walk(root.Command, command)

	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %d commands to %s\n", len(commands), out)
	return nil
}
