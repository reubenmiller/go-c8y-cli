package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
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
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	consol, err := n.factory.Console()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UTC().UnixNano())

	var iter iterator.Iterator
	_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
		Name:        "input",
		InputFilter: func(b []byte) bool { return true },
		Disabled:    inputIterators.PipeOptions.Disabled,
		Required:    false,
	})(cmd, inputIterators)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	switch v := input.(type) {
	case iterator.Iterator:
		iter = v
	default:
		// use a single input iterator
		iter = iterator.NewRepeatIterator("", 1)
	}

	var schema *jsonschema.Schema
	if n.schema != "" {
		if strings.HasPrefix(n.schema, "{") {
			// Load schema from
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

	writeOutput := func(isJSON bool, output []byte) {
		if isJSON {
			_ = n.factory.WriteJSONToConsole(cfg, cmd, "", output)
		} else {
			fmt.Fprintf(consol, "%s\n", output)
		}
	}

	totalErrors := 0
	var lastErr error
	for {
		err = nil
		input, _, inputErr := iter.GetNext()

		if inputErr == io.EOF {
			break
		}

		if totalErrors >= cfg.AbortOnErrorCount() {
			msg := fmt.Sprintf("Too many errors. total=%d, max=%d. lastError=%s", totalErrors, cfg.AbortOnErrorCount(), lastErr)
			return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAbortedWithErrors, msg)
		}

		isJSON := jsonUtilities.IsJSONObject(input)

		// schema match
		if schema != nil {
			var val interface{}
			_ = json.Unmarshal(input, &val)
			err = schema.Validate(val)

			if err == nil {
				writeOutput(isJSON, input)
			} else {
				err = fmt.Errorf("%w. input does not match json schema. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.schema)
			}
		}

		// regex match
		if pattern != nil {
			if err == nil {
				if !pattern.Match(input) {
					err = fmt.Errorf("%w. input does not match pattern. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.regex)
				} else {
					writeOutput(isJSON, input)
				}
			}
		}

		// exact match
		if err == nil {
			if n.exact != "" {
				if !bytes.Equal(input, []byte(n.exact)) {
					err = fmt.Errorf("%w. input does not match. got=%s, wanted=%s", cmderrors.ErrAssertion, input, n.exact)
				} else {
					writeOutput(isJSON, input)
				}
			}
		}

		if err != nil {
			if !errors.Is(err, cmderrors.ErrAssertion) || n.strictMode {
				totalErrors++
				lastErr = n.factory.CheckPostCommandError(err)

				// wrap error so it is not printed twice, and is still an assertion error
				cErr := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitAssertionError, lastErr)
				cErr.Processed = true
				lastErr = cErr
			}
		}
	}
	if totalErrors > 0 {
		return lastErr
	}
	return nil

}
