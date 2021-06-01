package docs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, func(s string, opts ...string) string { return s })
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string, ...string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	// buf.WriteString("## " + name + "\n\n")
	buf.WriteString(cmd.Short + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string, opts ...string) string { return s }
	emptyStr := func(s string, opts ...string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string, ...string) string) error {
	for _, c := range cmd.Commands() {
		_, forceGeneration := c.Annotations["markdown:generate"]
		if c.Hidden && !forceGeneration {
			continue
		}

		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	if basenameOverride, found := cmd.Annotations["markdown:basename"]; found {
		basename = basenameOverride + ".md"
	}

	docPath := []string{dir}

	rootCmdName := cmd.Root().CommandPath()
	pathWithoutBase := strings.Replace(cmd.CommandPath(), rootCmdName+" ", "", 1)

	commandParts := strings.Split(pathWithoutBase, " ")

	if len(commandParts) == 1 {
		commandParts = commandParts[0:]
	} else if len(commandParts) > 1 {
		if len(cmd.Commands()) > 0 {
			commandParts = commandParts[:]
		} else {
			commandParts = commandParts[0 : len(commandParts)-1]
		}
	}

	if len(commandParts) > 0 {
		docPath = append(docPath, commandParts...)
	}
	docPath = append(docPath, basename)
	filename := strings.Join(docPath, string(filepath.Separator))

	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	category := cmd.Use
	if cmd.HasSubCommands() {
		category = cmd.Use
	} else if cmd.HasParent() {
		category = cmd.Parent().Use
	}

	if _, err := io.WriteString(f, filePrepender(filename, category, cmd.Use, cmd.CommandPath())); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}
