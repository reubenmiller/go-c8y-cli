package powershell

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/clisurface"
	"github.com/spf13/cobra"
)

// GenerateFromTree projects the PSc8y cmdlets and tests from a live cobra
// command tree instead of the REST spec. It walks the CLI surface
// (clisurface.Walk) and, for each command, reconstructs the spec-shaped value
// the renderer expects and reuses the existing renderer (GenerateSpec). This is
// the inverted pipeline described in proposals/CLI_CODEGEN_INVERSION.md: the
// hand-written command tree is the source of truth, PowerShell is a projection.
//
// filter, when non-empty, limits generation to a command subtree by path
// prefix (e.g. "devices" or "devices get").
func GenerateFromTree(root *cobra.Command, filter string) ([]GeneratedFile, error) {
	var all []GeneratedFile
	for _, cmd := range clisurface.Walk(root, filter) {
		doc := surfaceToSpecDoc(cmd)
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		files, err := GenerateSpec(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", strings.Join(cmd.Path, " "), err)
		}
		all = append(all, files...)
	}
	return reduceFiles(all), nil
}

// surfaceToSpecDoc maps a neutral command surface to the single-command spec
// document the PowerShell renderer consumes.
func surfaceToSpecDoc(c clisurface.Command) *psSpecDoc {
	psName := c.PowershellName
	if psName == "" {
		psName = derivePowershellName(c.Noun, c.Name)
	}

	accept, itemType := c.Accept, c.ItemType
	// The renderer adds the "Collection" common parameter set when the accept
	// type contains "collection"; synthesize a marker when only the collection
	// item type or property is known.
	if itemType != "" && !strings.Contains(strings.ToLower(accept), "collection") {
		if accept == "" {
			accept = "application/json"
		}
		accept += "; collection"
	} else if accept == "" && c.CollectionProperty != "" {
		accept = "application/json; collection"
	}

	examples := make([]any, 0, len(c.Examples))
	for _, e := range c.Examples {
		examples = append(examples, e)
	}

	args := make([]psSpecArg, 0, len(c.Flags))
	for _, f := range c.Flags {
		arg := psSpecArg{
			Name:          f.Name,
			Type:          specType(f.Type, f.Pipeline),
			Description:   cleanUsage(f.Usage),
			Required:      f.Required,
			Pipeline:      f.Pipeline,
			ValidationSet: f.ValidateSet,
		}
		if f.Pipeline {
			pos := 0.0
			arg.Position = &pos
		}
		args = append(args, arg)
	}

	return &psSpecDoc{
		Group: psSpecGroup{Name: c.Noun},
		Commands: []psSpecCommand{{
			Name:            c.Name,
			Method:          c.Method,
			Description:     c.Short,
			DescriptionLong: c.Long,
			Accept:          accept,
			CollectionType:  itemType,
			Alias:           psSpecAlias{Go: c.Name, Powershell: psName},
			Examples:        psSpecExamples{Powershell: examples},
			QueryParameters: args,
		}},
	}
}

// specType maps a pflag value type to the spec argument type the renderer
// understands. Pipeline slice flags use "id" so they render as [object[]] with
// pipeline binding; scalar pipeline flags stay as their base type (the renderer
// upgrades string -> object[] for piped values).
func specType(pflagType string, isPipe bool) string {
	if isPipe && (pflagType == "stringSlice" || pflagType == "stringArray") {
		return "id"
	}
	switch pflagType {
	case "bool":
		return "boolean"
	case "stringSlice", "stringArray":
		return "string[]"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "count":
		return "integer"
	case "float32", "float64":
		return "float"
	default:
		// string, duration, ip and any custom value types degrade to a plain
		// string parameter, which is always a safe PowerShell binding.
		return "string"
	}
}

// cleanUsage strips the CLI-only hint suffixes from a flag usage string so the
// renderer (which re-adds " (required)") does not double them up.
func cleanUsage(usage string) string {
	for _, suffix := range []string{" (accepts pipeline)", " (required)"} {
		usage = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(usage), suffix))
	}
	return usage
}

// derivePowershellName builds a fallback cmdlet name from the command verb and
// noun when the command carries no explicit c8y:powershell.name annotation. It
// is intentionally simple; explicit annotations are preferred.
func derivePowershellName(noun, verb string) string {
	psVerb, collection := powershellVerb(verb)
	parts := strings.Fields(noun)
	if len(parts) == 0 {
		parts = []string{verb}
	}
	// PowerShell nouns are the full entity in CamelCase (e.g. "device groups"
	// -> "DeviceGroup"); only the last word is singularized.
	var b strings.Builder
	for i, p := range parts {
		if i == len(parts)-1 {
			p = singularize(p)
		}
		b.WriteString(upperFirst(p))
	}
	name := psVerb + "-" + b.String()
	if collection {
		name += "Collection"
	}
	return name
}

// powershellVerb maps a CLI verb to an approved PowerShell verb and reports
// whether the result is a collection (list) cmdlet.
func powershellVerb(verb string) (string, bool) {
	switch strings.ToLower(verb) {
	case "list":
		return "Get", true
	case "get":
		return "Get", false
	case "create", "new":
		return "New", false
	case "update", "set":
		return "Update", false
	case "delete", "remove":
		return "Remove", false
	default:
		return upperFirst(verb), false
	}
}

func singularize(s string) string {
	switch {
	case strings.HasSuffix(s, "ies") && len(s) > 3:
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(s, "ses") && len(s) > 3:
		return s[:len(s)-2]
	case strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") && len(s) > 1:
		return s[:len(s)-1]
	default:
		return s
	}
}

// psSpecDoc and friends mirror the subset of the JSON spec schema that the
// PowerShell renderer reads. Marshalling these and feeding GenerateSpec keeps
// the renderer (templates, argument handling, test logic) untouched.
type psSpecDoc struct {
	Group    psSpecGroup     `json:"group"`
	Commands []psSpecCommand `json:"commands"`
}

type psSpecGroup struct {
	Name string `json:"name"`
}

type psSpecCommand struct {
	Name            string         `json:"name"`
	Method          string         `json:"method"`
	Description     string         `json:"description"`
	DescriptionLong string         `json:"descriptionLong,omitempty"`
	Link            string         `json:"link,omitempty"`
	Accept          string         `json:"accept,omitempty"`
	CollectionType  string         `json:"collectionType,omitempty"`
	Alias           psSpecAlias    `json:"alias"`
	Examples        psSpecExamples `json:"examples"`
	QueryParameters []psSpecArg    `json:"queryParameters"`
}

type psSpecAlias struct {
	Go         string `json:"go"`
	Powershell string `json:"powershell"`
}

type psSpecExamples struct {
	// Powershell holds plain example strings (doc-only). Using []any keeps the
	// JSON shape (array of strings) the renderer's Value.Items() expects.
	Powershell []any `json:"powershell,omitempty"`
}

type psSpecArg struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Description   string   `json:"description,omitempty"`
	Required      bool     `json:"required,omitempty"`
	Pipeline      bool     `json:"pipeline,omitempty"`
	Position      *float64 `json:"position,omitempty"`
	ValidationSet []string `json:"validationSet,omitempty"`
}
