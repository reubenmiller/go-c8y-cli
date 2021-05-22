package cmdutil

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
)

func buildDummyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		PreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func Test_WithTemplateValue(t *testing.T) {
	cmd := buildDummyCommand()
	flags.WithOptions(
		cmd,
		flags.WithData(),
		flags.WithTemplate(),
	)
	inputIterator, _ := NewRequestInputIterators(cmd, nil)

	cmd.SetArgs([]string{"--template", "{value: input.index}"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	body := mapbuilder.NewInitializedMapBuilder()
	err := flags.WithBody(
		cmd,
		body,
		inputIterator,
		WithTemplateValue(nil),
		flags.WithTemplateVariablesValue(),
	)

	assert.OK(t, err)
	assert.EqualMarshalJSON(t, body, `{"value":"1"}`)
}
