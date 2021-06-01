package flags

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// GetIDs returns a list of IDs
// --id 1234,1234 							(comma seperated list (without spaces!))
// --id "22437097744 1235"					[22437097744, 1235]
// --id 22437097744 1235,1234				(requires positional arguments as well)
// --id 22437097744 1235,1234 asdfasdf		asdfasdf will be ignored as it does not match the pattern
func GetIDs(cmd *cobra.Command, args []string) (ids []string) {
	if values, err := cmd.Flags().GetStringArray("id"); err != nil {
		cmd.PrintErrf("Missing flag --id")
	} else {
		ids = GetIDArray(append(values, args...))
	}
	return
}

func GetIDArray(values []string) (ids []string) {
	for _, value := range values {
		parts := strings.Split(strings.ReplaceAll(value, " ", ","), ",")

		for _, part := range parts {
			// Only add uint looking values
			if _, err := strconv.ParseUint(part, 10, 64); part != "" && err == nil {
				ids = append(ids, part)
			}
		}
	}
	return
}
