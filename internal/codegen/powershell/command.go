package powershell

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/codegen"
)

// docBaseURL is where the cmdlet doc strings link to (the generated c8y
// native command documentation).
const docBaseURL = "https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y"

// renderCmdlet ports New-C8yApiPowershellCommand.ps1 (without the test file
// generation, see renderTest). It returns the content of the
// Public/<CmdletName>.ps1 file, excluding the BOM and trailing newline added
// when writing the file.
func renderCmdlet(command *codegen.Value, noun string) (string, error) {
	cmdletName := command.Get("alias").Get("powershell").Str()
	nounLower := strings.ReplaceAll(strings.ToLower(noun), "/", " ")
	resultType := command.Get("accept").Str()
	resultItemType := command.Get("collectionType").Str()
	verb := command.Get("alias").Get("go").Str()
	documentationLink := command.Get("link").Str()

	//
	// Meta information
	//
	synopsis := command.Get("description").Str()
	descriptionLong := command.Get("descriptionLong").Str()
	if descriptionLong == "" {
		descriptionLong = synopsis
	}

	var examples []string
	for _, example := range command.Get("examples").Get("powershell").Items() {
		if example.Get("command").Truthy() {
			examples = append(examples, fmt.Sprintf("PS> %s\n\n%s", example.Get("command").Str(), example.Get("description").Str()))
		} else {
			examples = append(examples, example.Str())
		}
	}

	doc := &strings.Builder{}
	if synopsis != "" {
		doc.WriteString(".SYNOPSIS\n")
		doc.WriteString(synopsis + "\n\n")
	}
	if descriptionLong != "" {
		doc.WriteString(".DESCRIPTION\n")
		doc.WriteString(descriptionLong + "\n\n")
	}

	// Link go command
	doc.WriteString(".LINK\n")
	pageName := strings.ReplaceAll(nounLower+"_"+verb, " ", "_")
	doc.WriteString(docBaseURL + "/" + pageName + "\n\n")

	//
	// Arguments
	//
	argumentSources := collectArgumentSources(command)

	var cmdletParameters []string
	iteratorVariable := ""

	for _, iArg := range argumentSources {
		name := iArg.Get("name").Str()
		alias := iArg.Get("alias").Str()
		readFromPipeline := iArg.Get("pipeline").Truthy() ||
			strings.EqualFold(name, "id") || strings.EqualFold(alias, "id")

		if iArg.Get("deprecated").Truthy() {
			continue
		}

		item, err := newCmdletArgument(name, iArg.Get("type").Str(), iArg.Get("description").Str(), iArg.Get("required").Str(), readFromPipeline)
		if err != nil {
			return "", fmt.Errorf("%s: %w", cmdletName, err)
		}

		if item.Ignore {
			continue
		}

		if readFromPipeline {
			if lower := strings.ToLower(item.Type); lower == "string" || lower == "long" {
				item.Type = "object[]"
			}
			iteratorVariable = "$" + item.Name
		}

		// Parameter definition
		param := &strings.Builder{}
		param.WriteString("        # " + item.Description + "\n")
		param.WriteString("        [Parameter(" + strings.Join(item.Definition, ",\n                   ") + ")]\n")

		if alias != "" {
			param.WriteString("        [Alias(\"" + alias + "\")]\n")
		}

		// Validate set
		if validationSet := iArg.Get("validationSet"); !validationSet.IsNull() {
			quoted := make([]string, 0)
			for _, v := range validationSet.Items() {
				quoted = append(quoted, "'"+v.Str()+"'")
			}
			param.WriteString("        [ValidateSet(" + strings.Join(quoted, ",") + ")]\n")
		}

		param.WriteString("        [" + item.Type + "]\n")
		param.WriteString("        $" + item.Name)
		cmdletParameters = append(cmdletParameters, param.String())
	}

	//
	// Add common parameters related to Method
	//
	var commonSetNames []string
	switch strings.ToUpper(command.Get("method").Str()) {
	case "GET":
		commonSetNames = append(commonSetNames, "Get")
	case "POST":
		commonSetNames = append(commonSetNames, "Create", "Template")
	case "PUT":
		commonSetNames = append(commonSetNames, "Update", "Template")
	case "DELETE":
		commonSetNames = append(commonSetNames, "Delete")
	}
	if strings.Contains(strings.ToLower(resultType), "collection") {
		commonSetNames = append(commonSetNames, "Collection")
	}

	// Examples
	for _, example := range examples {
		doc.WriteString(".EXAMPLE\n")
		doc.WriteString(example + "\n\n")
	}

	// Doc link
	if documentationLink != "" {
		doc.WriteString(".LINK " + documentationLink + "\n")
	}

	//
	// Template
	//
	b := &strings.Builder{}
	b.WriteString("# Code generated from specification version 1.0.0: DO NOT EDIT\n")
	b.WriteString("Function " + cmdletName + " {\n")
	b.WriteString("<#\n")
	b.WriteString(doc.String() + "\n")
	b.WriteString("#>\n")
	b.WriteString("    [cmdletbinding(PositionalBinding=$true,\n")
	b.WriteString("                   HelpUri='" + documentationLink + "')]\n")
	b.WriteString("    [Alias()]\n")
	b.WriteString("    [OutputType([object])]\n")
	b.WriteString("    Param(\n")
	b.WriteString(strings.Join(cmdletParameters, ",\n\n") + "\n")
	b.WriteString("    )\n")
	b.WriteString("    DynamicParam {\n")
	b.WriteString("        Get-ClientCommonParameters -Type \"" + strings.Join(commonSetNames, "\", \"") + "\"\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    Begin {\n")
	b.WriteString("\n")
	b.WriteString("        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {\n")
	b.WriteString("            # Inherit preference variables\n")
	b.WriteString("            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState\n")
	b.WriteString("        }\n")
	b.WriteString("\n")
	b.WriteString("        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command \"" + nounLower + " " + verb + "\"\n")
	b.WriteString("        $ClientOptions = Get-ClientOutputOption $PSBoundParameters\n")
	b.WriteString("        $TypeOptions = @{\n")
	b.WriteString("            Type = \"" + resultType + "\"\n")
	b.WriteString("            ItemType = \"" + resultItemType + "\"\n")
	b.WriteString("            BoundParameters = $PSBoundParameters\n")
	b.WriteString("        }\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    Process {\n")
	b.WriteString(renderProcessBody(nounLower, verb, iteratorVariable) + "\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    End {}\n")
	b.WriteString("}")

	return b.String(), nil
}

// collectArgumentSources gathers the cmdlet arguments from the specification
// sections in their CLI order: a stable sort by position (default 20) with
// skipped arguments and the special "data" argument removed. Children of
// query parameters replace their parent (excluding stringStatic entries).
func collectArgumentSources(command *codegen.Value) []*codegen.Value {
	var sources []*codegen.Value
	sources = append(sources, command.Get("pathParameters").Items()...)

	for _, item := range command.Get("queryParameters").Items() {
		if item.Get("children").Truthy() {
			// Ignore the item, and only use the children to build the cli arguments
			for _, child := range item.Get("children").Items() {
				if !child.Get("type").StrEquals("stringStatic") {
					sources = append(sources, child)
				}
			}
		} else {
			sources = append(sources, item)
		}
	}

	sources = append(sources, command.Get("body").Items()...)
	sources = append(sources, command.Get("headerParameters").Items()...)
	sources = append(sources, command.Get("options").Items()...)

	// (stable) sort argument sources by position to control the expected order on cli
	positions := make([]float64, len(sources))
	for i, source := range sources {
		positions[i] = 20
		if p := source.Get("position"); !p.IsNull() {
			if v, err := strconv.ParseFloat(p.Str(), 64); err == nil {
				positions[i] = v
			}
		}
	}
	order := make([]int, len(sources))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return positions[order[a]] < positions[order[b]]
	})

	filtered := make([]*codegen.Value, 0, len(sources))
	for _, idx := range order {
		source := sources[idx]
		if source.Get("skip").Truthy() {
			continue
		}
		if strings.EqualFold(source.Get("name").Str(), "data") {
			continue
		}
		filtered = append(filtered, source)
	}
	return filtered
}

// renderProcessBody ports the New-Body2 function. The confirmation statement
// branch is disabled in the original script, so the body only depends on
// whether a pipeline iterator variable exists.
func renderProcessBody(nounLower, verb, iteratorVariable string) string {
	c8yCommand := "c8y " + nounLower + " " + verb + " $c8yargs"

	b := &strings.Builder{}
	b.WriteString("\n")
	if iteratorVariable != "" {
		b.WriteString("        if ($ClientOptions.ConvertToPS) {\n")
		b.WriteString("            " + iteratorVariable + " `\n")
		b.WriteString("            | Group-ClientRequests `\n")
		b.WriteString("            | " + c8yCommand + " `\n")
		b.WriteString("            | ConvertFrom-ClientOutput @TypeOptions\n")
		b.WriteString("        }\n")
		b.WriteString("        else {\n")
		b.WriteString("            " + iteratorVariable + " `\n")
		b.WriteString("            | Group-ClientRequests `\n")
		b.WriteString("            | " + c8yCommand + "\n")
		b.WriteString("        }\n")
		b.WriteString("        ")
	} else {
		b.WriteString("        if ($ClientOptions.ConvertToPS) {\n")
		b.WriteString("            " + c8yCommand + " `\n")
		b.WriteString("            | ConvertFrom-ClientOutput @TypeOptions\n")
		b.WriteString("        }\n")
		b.WriteString("        else {\n")
		b.WriteString("            " + c8yCommand + "\n")
		b.WriteString("        }")
	}
	return b.String()
}
