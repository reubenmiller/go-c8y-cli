package cmdutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	testify_assert "github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

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
	inputIterators, _ := NewRequestInputIterators(cmd, nil)
	err := flags.WithHeaders(
		cmd,
		header,
		inputIterators,
		flags.WithIntValue("count", "CountValue"),
		flags.WithBoolValue("csv", "Accept", "text/csv"),
		flags.WithStringValue("type"),
		flags.WithStringValue("type", "Content-Type", "text/%s"),
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

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("nextID", "nextID", true),
	)

	cmd.SetArgs([]string{"--nextID", "7777,8888", "--editable", "--type", "myType", "--dateFrom", "-1d", "--newChild", "1111,2222"})
	cmdErr := cmd.Execute()
	assert.OK(t, cmdErr)

	inputIterators, _ := NewRequestInputIterators(cmd, nil)
	body := mapbuilder.NewInitializedMapBuilder(true)
	err := flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithIntValue("count", "CountValue"),
		flags.WithBoolValue("editable"),
		flags.WithStringValue("type"),
		flags.WithStringValue("type", "typeMapping", "text/%s"),
		flags.WithRelativeTimestamp("time"),
		flags.WithRelativeTimestamp("dateFrom"),
		flags.WithRelativeTimestamp("dateFrom", "dateTo"),
		flags.WithStringSliceValues("newChild", "managedObject.id", ""),
		flags.WithStringValue("nextID", "nextID", ""),
	)
	assert.OK(t, err)

	// The body is built as json (since the map builder stores the raw json),
	// and the iterator (nextID) gives the next value each time it is marshalled
	body1, err := body.MarshalJSON()
	assert.OK(t, err)
	time.Sleep(1 * time.Millisecond)
	body2, err := body.MarshalJSON()
	assert.OK(t, err)

	assert.True(t, gjson.GetBytes(body1, "CountValue").Raw == "2")
	assert.True(t, gjson.GetBytes(body1, "editable").Raw == "true")
	assert.True(t, gjson.GetBytes(body1, "type").String() == "myType")
	assert.True(t, gjson.GetBytes(body1, "typeMapping").String() == "text/myType")
	assert.True(t, gjson.GetBytes(body1, "managedObject.id").Raw == `["1111","2222"]`)

	// iterator check
	assert.True(t, gjson.GetBytes(body1, "nextID").String() == "7777")
	assert.True(t, gjson.GetBytes(body2, "nextID").String() == "8888")

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

	inputIterators, _ := NewRequestInputIterators(cmd, nil)
	query := flags.NewQueryTemplate()
	err := flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithIntValue("count", "CountValue"),
		flags.WithBoolValue("editable"),
		flags.WithStringValue("type"),
		flags.WithStringValue("type", "typeMapping", "text/%s"),
		flags.WithRelativeTimestamp("dateFrom"),
		flags.WithRelativeTimestamp("dateFrom", "dateTo"),
	)
	assert.OK(t, err)

	queryValue, _, err := query.Execute(false)
	assert.OK(t, err)
	assert.True(t, queryValue.Get("CountValue") == "2")
	assert.True(t, queryValue.Get("editable") == "true")
	assert.True(t, queryValue.Get("type") == "myType")
	assert.True(t, queryValue.Get("typeMapping") == "text/myType")
}

func Test_PathEscaping(t *testing.T) {
	cmd := buildDummyCommand()

	cmd.Flags().String("category", "", "String type")
	cmd.Flags().String("name", "", "String type")

	cases := []struct {
		Args     []string
		Expected string
	}{
		{
			Args:     []string{"--category", "foo#bar", "--name", "bar"},
			Expected: "foo%23bar/bar",
		},
		{
			Args:     []string{"--category", "foo%23bar", "--name", "bar"},
			Expected: "foo%23bar/bar",
		},
	}

	for _, testcase := range cases {
		cmd.SetArgs(testcase.Args)
		cmdErr := cmd.Execute()
		testify_assert.NoError(t, cmdErr)

		inputIterators, _ := NewRequestInputIterators(cmd, nil)
		path := flags.NewStringTemplate("{category}/{name}")
		err := flags.WithPathParameters(
			cmd,
			path,
			inputIterators,
			flags.WithStringValue("category"),
			flags.WithStringValue("name"),
		)
		testify_assert.NoError(t, err)

		pathValue, _, err := path.Execute(false)
		testify_assert.NoError(t, err)
		testify_assert.Equal(t, testcase.Expected, pathValue)
	}
}
