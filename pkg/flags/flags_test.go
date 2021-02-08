package flags

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
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

func Test_HeaderFlags(t *testing.T) {
	cmd := buildDummyCommand()

	cmd.Flags().Int("count", 2, "Integer type")
	cmd.Flags().String("type", "", "String type")
	cmd.Flags().String("dateFrom", "", "Relative date")
	cmd.Flags().Bool("csv", false, "Boolean type")

	cmd.SetArgs([]string{"--csv", "--type", "myType", "--dateFrom", "0s"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	header := http.Header{}
	err := WithHeaders(
		cmd,
		header,
		WithIntValue("count", "CountValue"),
		WithBoolValue("csv", "Accept", "text/csv"),
		WithStringValue("type"),
		WithStringValue("type", "Content-Type", "text/%s"),
	)
	assert.OK(t, err)
	assert.True(t, header.Get("Accept") == "text/csv")
	assert.True(t, header.Get("type") == "myType")
	assert.True(t, header.Get("Content-Type") == "text/myType")
	assert.True(t, header.Get("CountValue") == "2")
}

func Test_Body(t *testing.T) {
	cmd := buildDummyCommand()

	cmd.Flags().Int("count", 2, "Integer type")
	cmd.Flags().String("type", "", "String type")
	cmd.Flags().String("dateFrom", "0s", "Relative date")
	cmd.Flags().Bool("editable", false, "Boolean type")
	cmd.Flags().StringSlice("newChild", []string{""}, "dummy child reference")

	cmd.SetArgs([]string{"--editable", "--type", "myType", "--dateFrom", "-1d", "--newChild", "12345"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	body := mapbuilder.NewInitializedMapBuilder()
	err := WithBody(
		cmd,
		body,
		WithIntValue("count", "CountValue"),
		WithBoolValue("editable"),
		WithStringValue("type"),
		WithStringValue("type", "typeMapping", "text/%s"),
		WithRelativeTimestamp("dateFrom"),
		WithRelativeTimestamp("dateFrom", "dateTo"),
		WithStringSliceValues("newChild", "managedObject.id", ""),
	)
	assert.OK(t, err)

	bodyMap := body.GetMap()
	assert.True(t, bodyMap["CountValue"].(int) == 2)
	assert.True(t, bodyMap["editable"].(bool) == true)
	assert.True(t, bodyMap["type"].(string) == "myType")
	assert.True(t, bodyMap["typeMapping"].(string) == "text/myType")
	childIds := bodyMap["managedObject"].(map[string]interface{})["id"].([]string)
	assert.True(t, childIds[0] == "12345")
	assert.True(t, len(childIds) == 0)
}

func Test_QueryParameters(t *testing.T) {
	cmd := buildDummyCommand()

	cmd.Flags().Int("count", 2, "Integer type")
	cmd.Flags().String("type", "", "String type")
	cmd.Flags().String("dateFrom", "0s", "Relative date")
	cmd.Flags().Bool("editable", false, "Boolean type")

	cmd.SetArgs([]string{"--editable", "--type", "myType", "--dateFrom", "-1d"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	query := flags.NewQueryTemplate()
	err := WithQueryParameters(
		cmd,
		query,
		WithIntValue("count", "CountValue"),
		WithBoolValue("editable"),
		WithStringValue("type"),
		WithStringValue("type", "typeMapping", "text/%s"),
		WithRelativeTimestamp("dateFrom"),
		WithRelativeTimestamp("dateFrom", "dateTo"),
	)
	assert.OK(t, err)

	assert.True(t, query.Get("CountValue") == "2")
	assert.True(t, query.Get("editable") == "true")
	assert.True(t, query.Get("type") == "myType")
	assert.True(t, query.Get("typeMapping") == url.QueryEscape("text/myType"))
}
