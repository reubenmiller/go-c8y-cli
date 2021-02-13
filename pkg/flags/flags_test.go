package flags

import (
	"net/http"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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
	inputIterators, _ := NewRequestInputIterators(cmd)
	err := WithHeaders(
		cmd,
		header,
		inputIterators,
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
	cmd.Flags().String("time", "0s", "Relative time")
	cmd.Flags().String("dateFrom", "0s", "Relative date")
	cmd.Flags().Bool("editable", false, "Boolean type")
	cmd.Flags().StringSlice("newChild", []string{""}, "dummy child reference")
	cmd.Flags().StringSlice("nextID", []string{""}, "dummy child reference")

	WithOptions(
		cmd,
		WithExtendedPipelineSupport("nextID", "nextID", true),
	)

	cmd.SetArgs([]string{"--nextID", "7777,8888", "--editable", "--type", "myType", "--dateFrom", "-1d", "--newChild", "1111,2222"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	inputIterators, _ := NewRequestInputIterators(cmd)
	body := mapbuilder.NewInitializedMapBuilder()
	err := WithBody(
		cmd,
		body,
		inputIterators,
		WithIntValue("count", "CountValue"),
		WithBoolValue("editable"),
		WithStringValue("type"),
		WithStringValue("type", "typeMapping", "text/%s"),
		WithRelativeTimestamp("time"),
		WithRelativeTimestamp("dateFrom"),
		WithRelativeTimestamp("dateFrom", "dateTo"),
		WithStringSliceValues("newChild", "managedObject.id", ""),
		WithStringValue("nextID", "nextID", ""),
	)
	assert.OK(t, err)

	bodyMap := body.GetMap()
	assert.True(t, bodyMap["CountValue"].(int) == 2)
	assert.True(t, bodyMap["editable"].(bool) == true)
	assert.True(t, bodyMap["type"].(string) == "myType")
	assert.True(t, bodyMap["typeMapping"].(string) == "text/myType")

	childIds := bodyMap["managedObject"].(map[string]interface{})["id"].([]string)
	assert.True(t, childIds[0] == "1111")
	assert.True(t, len(childIds) == 1)

	// iterator check
	nextID := bodyMap["nextID"].(iterator.Iterator)
	assert.True(t, nextID != nil)

	body1, err := body.MarshalJSON()
	assert.OK(t, err)
	time.Sleep(1 * time.Millisecond)
	body2, err := body.MarshalJSON()
	assert.OK(t, err)

	time1 := gjson.GetBytes(body1, "time").String()
	time2 := gjson.GetBytes(body2, "time").String()
	assert.True(t, time1 != time2)

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

	inputIterators, _ := NewRequestInputIterators(cmd)
	query := NewQueryTemplate()
	err := WithQueryParameters(
		cmd,
		query,
		inputIterators,
		WithIntValue("count", "CountValue"),
		WithBoolValue("editable"),
		WithStringValue("type"),
		WithStringValue("type", "typeMapping", "text/%s"),
		WithRelativeTimestamp("dateFrom"),
		WithRelativeTimestamp("dateFrom", "dateTo"),
	)
	assert.OK(t, err)

	queryValue, err := query.Execute(false)
	assert.OK(t, err)
	assert.True(t, queryValue.Get("CountValue") == "2")
	assert.True(t, queryValue.Get("editable") == "true")
	assert.True(t, queryValue.Get("type") == "myType")
	assert.True(t, queryValue.Get("typeMapping") == "text/myType")
}
