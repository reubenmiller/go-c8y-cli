package format

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdFormat struct {
	*subcommand.SubCommand

	factory          *cmdutil.Factory
	verifyIdempotent bool
	inPlace          bool
	backup           bool
	indent           int
	addComment       bool
}

func NewCmdFormat(f *cmdutil.Factory) *CmdFormat {
	ccmd := &CmdFormat{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "format",
		Short: "Format shell scripts",
		Long:  `Format shell scripts by adjusting -n flags for c8y commands`,
		Example: heredoc.Doc(`
			$ c8y scripts format script.sh
			$ c8y scripts format --in-place script.sh
			$ c8y scripts format --add-comment --indent 4 script.sh
		`),
		Args: func(cmd *cobra.Command, args []string) error {
			isExamples, _ := cmd.Root().PersistentFlags().GetBool("examples")
			if isExamples {
				return nil
			}
			return cobra.MinimumNArgs(1)(cmd, args)
		},
		RunE: ccmd.run,
	}

	cmd.Flags().BoolVar(&ccmd.verifyIdempotent, "verify-idempotent", false, "Verify that the transformation is idempotent")
	cmd.Flags().BoolVarP(&ccmd.inPlace, "in-place", "i", false, "Modify files in place (implies --verify-idempotent)")
	cmd.Flags().BoolVar(&ccmd.backup, "backup", false, "Create backup files before in-place modification")
	cmd.Flags().IntVar(&ccmd.indent, "indent", 0, "Number of spaces to use for indentation. If 0, then tabs will be used")
	cmd.Flags().BoolVar(&ccmd.addComment, "add-comment", false, "Add a comment block at the top explaining the script modifications")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdFormat) run(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// disable standard input
	cfg.DisableStdin()

	for _, inputFile := range args {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return err
		}

		opts := FormatOptions{
			AddComment: n.addComment,
			Indent:     n.indent,
		}

		firstOutput, err := FormatScript(string(data), opts)
		if err != nil {
			return err
		}

		if n.inPlace {
			if n.verifyIdempotent {
				// Verify idempotent for in-place changes
				secondOutput, err := FormatScript(firstOutput, opts)
				if err != nil {
					return err
				}
				if firstOutput != secondOutput {
					return fmt.Errorf("transformation is not idempotent for file %s: cannot modify in place", inputFile)
				}
			}
			// Create backup if requested
			if n.backup {
				backupFile := inputFile + ".bak"
				if err := copyFile(inputFile, backupFile); err != nil {
					return fmt.Errorf("failed to create backup for %s: %w", inputFile, err)
				}
			}
			// Write the modified content back to the file
			if err := os.WriteFile(inputFile, []byte(firstOutput), 0644); err != nil {
				return fmt.Errorf("failed to write to %s: %w", inputFile, err)
			}
		} else {
			if n.verifyIdempotent {
				// Verify idempotent
				secondOutput, err := FormatScript(firstOutput, opts)
				if err != nil {
					return err
				}
				if firstOutput != secondOutput {
					return fmt.Errorf("transformation is not idempotent for file %s: output changed on second run", inputFile)
				}
			}
			// Print to stdout
			if _, err := os.Stdout.WriteString(firstOutput); err != nil {
				return err
			}
		}
	}
	return nil
}

// FormatOptions represents the options for formatting
type FormatOptions struct {
	AddComment bool
	Indent     int
}

// FormatScript formats a shell script by adjusting -n flags for c8y commands
func FormatScript(input string, opts FormatOptions) (string, error) {
	parser := syntax.NewParser(syntax.KeepComments(true))
	file, err := parser.Parse(strings.NewReader(input), "")
	if err != nil {
		return "", err
	}

	type parentInfo struct {
		inPipeline      bool
		stdinRedirected bool
		isolatedContext bool // true when in if conditions or command substitutions where pipelines don't affect -n removal
		isPipelineStart bool // true for the first command in a pipeline
	}

	// Helper: recursively walk with parent context
	var walk func(node syntax.Node, parent syntax.Node, info parentInfo)
	walk = func(node syntax.Node, parent syntax.Node, info parentInfo) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *syntax.BinaryCmd:
			if n.Op == syntax.Pipe {
				if n.X != nil {
					walk(n.X, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: true})
				}
				if n.Y != nil {
					walk(n.Y, n, parentInfo{inPipeline: true, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: false})
				}
			} else {
				if n.X != nil {
					walk(n.X, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: false})
				}
				if n.Y != nil {
					walk(n.Y, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: false})
				}
			}
		case *syntax.Redirect:
			if n.N != nil {
				if n.Op == syntax.RdrIn {
					walk(n.N, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: true, isolatedContext: info.isolatedContext, isPipelineStart: info.isPipelineStart})
				} else {
					walk(n.N, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: info.isPipelineStart})
				}
			}
		case *syntax.CmdSubst:
			// Command substitution $(...) - commands inside should keep -n
			for _, stmt := range n.Stmts {
				walk(stmt, n, parentInfo{inPipeline: false, stdinRedirected: info.stdinRedirected, isolatedContext: true, isPipelineStart: false})
			}
		case *syntax.IfClause:
			// If condition - commands in condition should keep -n
			for _, stmt := range n.Cond {
				walk(stmt, n, parentInfo{inPipeline: false, stdinRedirected: info.stdinRedirected, isolatedContext: true, isPipelineStart: false})
			}
			for _, stmt := range n.Then {
				walk(stmt, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: info.isPipelineStart})
			}
			if n.Else != nil {
				walk(n.Else, n, parentInfo{inPipeline: info.inPipeline, stdinRedirected: info.stdinRedirected, isolatedContext: info.isolatedContext, isPipelineStart: info.isPipelineStart})
			}
		case *syntax.CallExpr: // Walk assignment values
			for _, a := range n.Assigns {
				walk(a.Value, n, info)
			} // Only process if first arg is 'c8y'
			if len(n.Args) == 0 || len(n.Args[0].Parts) == 0 {
				return
			}
			word, ok := n.Args[0].Parts[0].(*syntax.Lit)
			if !ok || word.Value != "c8y" {
				return
			}
			// Find if -n is present and its index
			nIndex := -1
			for i, arg := range n.Args[1:] {
				if len(arg.Parts) == 0 {
					continue
				}
				if lit, ok := arg.Parts[0].(*syntax.Lit); ok && lit.Value == "-n" {
					nIndex = i + 1 // since n.Args[1:] , so index in n.Args is i+1
					break
				}
			}
			if info.isPipelineStart {
				// Add -n if not present (first command in pipeline should have -n)
				if nIndex < 0 {
					n.Args = append(n.Args[:1], append([]*syntax.Word{litWord("-n")}, n.Args[1:]...)...)
				}
			} else if info.inPipeline && !info.isolatedContext || info.stdinRedirected {
				// Remove -n if present
				if nIndex >= 0 {
					n.Args = append(n.Args[:nIndex], n.Args[nIndex+1:]...)
				}
			} else {
				// Add -n if not present
				if nIndex < 0 {
					n.Args = append(n.Args[:1], append([]*syntax.Word{litWord("-n")}, n.Args[1:]...)...)
				}
			}
			return
		case *syntax.Stmt:
			// Check if this statement has input redirects and the command is a simple call
			hasInputRedirect := false
			if _, ok := n.Cmd.(*syntax.CallExpr); ok {
				for _, r := range n.Redirs {
					if r.Op == syntax.RdrIn {
						hasInputRedirect = true
						break
					}
				}
			}
			newInfo := info
			if hasInputRedirect {
				newInfo.stdinRedirected = true
			}
			if n.Cmd != nil {
				walk(n.Cmd, n, newInfo)
			}
			for _, r := range n.Redirs {
				if r != nil {
					walk(r, n, info)
				}
			}
		default:
			syntax.Walk(n, func(child syntax.Node) bool {
				if child == node {
					return true
				}
				if child != nil {
					walk(child, node, info)
				}
				return false
			})
		}
	}

	walk(file, nil, parentInfo{})

	var buf bytes.Buffer
	printer := syntax.NewPrinter(syntax.Indent(uint(opts.Indent)), syntax.BinaryNextLine(false), syntax.Minify(false))
	if err := printer.Print(&buf, file); err != nil {
		return "", err
	}
	firstOutput := buf.String()

	if opts.AddComment {
		// Check if comment already exists
		if !strings.Contains(firstOutput, "# NOTE: This script has been processed by go-c8y-cli") {
			comment := `# -----------------------------------------------------------------------------------------------------
# NOTE: This script has been processed by go-c8y-cli
#
# The following modifications were made:
# - Commands in pipelines have -n flags removed as this would prevent the command receiving piped input
# - Commands not in pipelines have -n flags added to prevent issues in CI/CD environments or in
#   'while read' loops
# -----------------------------------------------------------------------------------------------------
`
			if strings.HasPrefix(firstOutput, "#!") {
				lines := strings.SplitN(firstOutput, "\n", 2)
				firstOutput = lines[0] + "\n" + comment + "\n" + lines[1]
			} else {
				firstOutput = comment + "\n" + firstOutput
			}
		}
	}

	return firstOutput, nil
}

// litWord creates a *syntax.Word from a literal string
func litWord(s string) *syntax.Word {
	return &syntax.Word{Parts: []syntax.WordPart{&syntax.Lit{Value: s}}}
}

// copyFile copies src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
