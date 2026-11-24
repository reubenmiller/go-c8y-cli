package json

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"
)

type CmdText struct {
	*subcommand.SubCommand

	schema     string
	exact      string
	regex      string
	strictMode bool
	factory    *cmdutil.Factory
}

func NewCmdText(f *cmdutil.Factory) *CmdText {
	ccmd := &CmdText{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Assert text",
		Long:  `Assert text input`,
		Example: heredoc.Doc(`
			$ echo "{\"name\": \"device01\"}" | c8y assert text --schema "myschema.json"
			Assert that the input matches a given schema

			$ echo "myname" | c8y assert text --exact "myname"
			Assert that the input text matches a given string exactly (case sensitive)

			$ echo "myname" | c8y assert text --regex '.*name$'
			Assert that the input text matches a given regular expression (case sensitive)

		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().String("input", "", "input value to be repeated (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.schema, "schema", "", "Match against a json schema")
	cmd.Flags().StringVar(&ccmd.exact, "exact", "", "Match exact text (case sensitive)")
	cmd.Flags().StringVar(&ccmd.regex, "regex", "", "Match against a regular expression (case sensitive)")
	cmd.Flags().BoolVar(&ccmd.strictMode, "strict", false, "Strict mode, fail if no match is found")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("input", "input", false),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdText) RunE(cmd *cobra.Command, args []string) error {
	var schema *jsonschema.Schema
	var err error
	if n.schema != "" {
		if strings.HasPrefix(n.schema, "{") {
			schema, err = jsonschema.CompileString("string://schema.json", n.schema)
		} else {
			schema, err = jsonschema.Compile(n.schema)
		}
		if err != nil {
			return err
		}
	}

	var pattern *regexp.Regexp
	if n.regex != "" {
		pattern, err = regexp.Compile("(?ms)" + n.regex)
		if err != nil {
			return err
		}
	}

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	err = r.InputPipeline(flags.PipelineOptions{
		Name:        "input",
		InputFilter: func(b []byte) bool { return true },
	})
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	r.Assert(n.strictMode)

	// No client: each item is checked and passed on as it is.
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		input := in.Input()
		return func(context.Context) any {
			if err := n.check(input, schema, pattern); err != nil {
				return err
			}
			if schema == nil && pattern == nil && n.exact == "" {
				return nil
			}
			return c8ystream.FromDocs(jsondoc.New(input))
		}, nil
	})
}

// check reports an assertion error when the input does not match all of the
// given conditions.
func (n *CmdText) check(input []byte, schema *jsonschema.Schema, pattern *regexp.Regexp) error {
	if schema != nil {
		var val any
		_ = json.Unmarshal(input, &val)
		if err := schema.Validate(val); err != nil {
			return fmt.Errorf("%w. input does not match json schema. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.schema)
		}
	}
	if pattern != nil && !pattern.Match(input) {
		return fmt.Errorf("%w. input does not match pattern. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.regex)
	}
	if n.exact != "" && !bytes.Equal(input, []byte(n.exact)) {
		return fmt.Errorf("%w. input does not match. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.exact)
	}
	return nil
}
