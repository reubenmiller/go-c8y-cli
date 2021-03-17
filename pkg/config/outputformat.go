package config

import "strings"

type OutputFormat int

const (
	// OutputJSON json output
	OutputJSON OutputFormat = iota

	// OutputTable table output
	OutputTable

	// OutputCSV csv output
	OutputCSV

	// OutputCSVWithHeader csv output with header
	OutputCSVWithHeader
)

func (f OutputFormat) String() string {
	values := map[OutputFormat]string{
		OutputJSON:          "json",
		OutputTable:         "table",
		OutputCSV:           "csv",
		OutputCSVWithHeader: "csvheader",
	}

	if v, ok := values[f]; ok {
		return v
	}
	return ""
}

func (f OutputFormat) FromString(name string) OutputFormat {
	values := map[string]OutputFormat{
		"json":      OutputJSON,
		"table":     OutputTable,
		"csv":       OutputCSV,
		"csvheader": OutputCSVWithHeader,
	}

	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return f
}
