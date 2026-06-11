package jsonfilter

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

func Test_MergeKeyGroups(t *testing.T) {
	// patterns which did not resolve against the first row should be
	// replaced by the keys resolved from later rows
	columns := MergeKeyGroups([][]KeyGroup{
		{
			{Pattern: "name", Keys: []string{"name"}},
			{Pattern: "c8y_firmware.versio*", Keys: nil},
		},
		{
			{Pattern: "name", Keys: []string{"name"}},
			{Pattern: "c8y_firmware.versio*", Keys: []string{"c8y_Firmware.version"}},
		},
	})
	assert.EqualMarshalJSON(t, columns, `["name","c8y_Firmware.version"]`)

	// keys are deduplicated case insensitively, and patterns which never
	// resolve are included as-is
	columns = MergeKeyGroups([][]KeyGroup{
		{
			{Pattern: "c8y_firmware.*", Keys: []string{"c8y_Firmware.version"}},
			{Pattern: "missing", Keys: nil},
		},
		{
			{Pattern: "c8y_firmware.*", Keys: []string{"c8y_firmware.VERSION", "c8y_Firmware.url"}},
			{Pattern: "missing", Keys: nil},
		},
	})
	assert.EqualMarshalJSON(t, columns, `["c8y_Firmware.version","c8y_Firmware.url","missing"]`)
}
