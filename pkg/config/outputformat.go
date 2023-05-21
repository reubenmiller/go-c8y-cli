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

	// OutputTSV tsv output
	OutputTSV

	// OutputCompletion completion output format, e.g. {value}\t{key1}: {value1} | {key2}: {value2}
	OutputCompletion

	// OutputCSVWithHeader csv output with header
	OutputCSVWithHeader

	// OutputServerResponse unparsed output as received from the server
	OutputServerResponse
)

func (f OutputFormat) String() string {
	values := map[OutputFormat]string{
		OutputJSON:           "json",
		OutputTable:          "table",
		OutputCSV:            "csv",
		OutputTSV:            "tsv",
		OutputCompletion:     "completion",
		OutputCSVWithHeader:  "csvheader",
		OutputServerResponse: "serverresponse",
	}

	if v, ok := values[f]; ok {
		return v
	}
	return ""
}

func (f OutputFormat) FromString(name string) OutputFormat {
	values := map[string]OutputFormat{
		"json":           OutputJSON,
		"table":          OutputTable,
		"csv":            OutputCSV,
		"tsv":            OutputTSV,
		"completion":     OutputCompletion,
		"csvheader":      OutputCSVWithHeader,
		"serverresponse": OutputServerResponse,
	}

	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return f
}
