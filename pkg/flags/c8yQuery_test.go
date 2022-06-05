package flags

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/spf13/cobra"
)

func Test_RelativeTimeIteratorWithFormatter(t *testing.T) {

	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().String("name1", "", "")
	cmd.Flags().String("name2", "", "")
	cmd.Flags().Bool("bool1", false, "")
	cmd.Flags().Bool("bool2", false, "")
	cmd.SetArgs([]string{"dummy", "--name1", "value1", "--bool2"})
	cmd.Execute()

	c8yQueryParts, err := WithC8YQueryOptions(
		cmd,
		nil,
		WithStaticStringValue("agent", "(has(com_cumulocity_model_Agent))"),
		WithStringValue("name1", "name", "(name eq '%s')"),
		WithStringValue("name2", "name", "(name eq '%s')"),
		WithBoolValue("bool1", "bool", "has(value1)"),
		WithBoolValue("bool2", "bool", "has(value2)"),
	)

	assert.OK(t, err)

	output := strings.Join(c8yQueryParts, " and ")
	assert.Equal(t, output, "(has(com_cumulocity_model_Agent)) and (name eq 'value1') and has(value2)")
}
