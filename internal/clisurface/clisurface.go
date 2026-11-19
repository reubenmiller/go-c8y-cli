// Package clisurface extracts the projection-neutral "CLI surface" from a live
// c8y cobra command tree: the commands, their flags, and the metadata carried
// on each (pipeline, validation sets, output type, powershell name, …). It is
// the single source that downstream projectors (PowerShell cmdlets, docs,
// tests, the surface manifest) walk, so they never diverge on what the tree
// contains. See proposals/CLI_CODEGEN_INVERSION.md.
package clisurface

import (
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command is one projectable leaf command and its surface. Field types stay
// projection-neutral: Flag.Type is the raw pflag type, descriptions are raw —
// each projector applies its own naming/type conventions.
type Command struct {
	// Path is the command path relative to the root, e.g. ["devices","get"].
	Path []string `json:"path"`
	// Name is the leaf verb (the command's Use), e.g. "get".
	Name string `json:"name"`
	// Noun is the space-joined parent path, e.g. "devices" or "devices user".
	Noun string `json:"noun"`

	Short    string   `json:"short,omitempty"`
	Long     string   `json:"long,omitempty"`
	Examples []string `json:"examples,omitempty"`

	// Method is the semantic REST method (from the semanticMethod annotation if
	// present, otherwise derived from the verb). Used to group common parameter
	// sets in projections.
	Method string `json:"method"`

	// Accept / ItemType are the response and collection-item media-types the
	// command produces (from the output-type annotation). Empty when unset.
	Accept   string `json:"accept,omitempty"`
	ItemType string `json:"itemType,omitempty"`

	// CollectionProperty is the default json property plucked from a collection
	// response (from the collectionProperty annotation).
	CollectionProperty string `json:"collectionProperty,omitempty"`

	// PowershellName is the explicit PowerShell cmdlet name (annotation), or
	// empty when the projector should derive one.
	PowershellName string `json:"powershellName,omitempty"`

	// PipelineFlag is the name of the flag that binds piped input, if any.
	PipelineFlag string `json:"pipelineFlag,omitempty"`

	Flags []Flag `json:"flags"`
}

// Flag is one non-inherited command flag and its surface metadata.
type Flag struct {
	Name string `json:"name"`
	// Type is the raw pflag value type, e.g. "string", "bool", "stringSlice".
	Type        string   `json:"type"`
	Usage       string   `json:"usage,omitempty"`
	Pipeline    bool     `json:"pipeline,omitempty"`
	Required    bool     `json:"required,omitempty"`
	ValidateSet []string `json:"validateSet,omitempty"`
}

// commonDynamicFlags are the local flags injected by client common-parameter
// machinery (data/template/processing mode). They are part of the invocation
// but not part of a command's own surface, so they are excluded.
var commonDynamicFlags = map[string]bool{
	flags.FlagDataName:                  true, // data
	flags.FlagDataTemplateName:          true, // template
	flags.FlagDataTemplateVariablesName: true, // templateVars
	flags.FlagProcessingModeName:        true, // processingMode
}

// Walk returns the surface of every projectable command under root. When filter
// is non-empty it limits the result to a command subtree by path prefix (e.g.
// "devices" or "devices get"); the root name prefix is optional.
func Walk(root *cobra.Command, filter string) []Command {
	filter = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(filter), root.Name()+" "))

	var out []Command
	var visit func(cmd *cobra.Command)
	visit = func(cmd *cobra.Command) {
		for _, child := range cmd.Commands() {
			visit(child)
		}
		if !Projectable(cmd) {
			return
		}
		if filter != "" && !matchesFilter(cmd, root, filter) {
			return
		}
		out = append(out, describe(cmd, root))
	}
	visit(root)
	return out
}

// Projectable reports whether a command should appear in the surface: a
// runnable leaf operation, not the root, not hidden, and not a built-in helper
// (help / completion / shell-completion shims).
func Projectable(cmd *cobra.Command) bool {
	if !cmd.HasParent() || !cmd.Runnable() || cmd.Hidden {
		return false
	}
	if cmd.IsAdditionalHelpTopicCommand() {
		return false
	}
	switch cmd.Name() {
	case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}
	return true
}

func matchesFilter(cmd, root *cobra.Command, filter string) bool {
	rel := strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), root.Name()))
	return rel == filter || strings.HasPrefix(rel, filter+" ")
}

func describe(cmd, root *cobra.Command) Command {
	tokens := strings.Fields(strings.TrimPrefix(cmd.CommandPath(), root.Name()+" "))
	verb := ""
	if len(tokens) > 0 {
		verb = tokens[len(tokens)-1]
	}
	noun := ""
	if len(tokens) > 1 {
		noun = strings.Join(tokens[:len(tokens)-1], " ")
	}

	method := strings.ToUpper(flags.GetSemanticMethodFromAnnotation(cmd))
	if method == "" {
		method = DeriveMethod(verb)
	}

	accept, itemType := flags.GetOutputTypeFromAnnotation(cmd)

	long := strings.TrimSpace(cmd.Long)
	if long == strings.TrimSpace(cmd.Short) {
		long = ""
	}

	return Command{
		Path:               tokens,
		Name:               verb,
		Noun:               noun,
		Short:              strings.TrimSpace(cmd.Short),
		Long:               long,
		Examples:           splitExamples(cmd.Example),
		Method:             method,
		Accept:             accept,
		ItemType:           itemType,
		CollectionProperty: flags.GetCollectionPropertyFromAnnotation(cmd),
		PowershellName:     flags.GetPowershellNameFromAnnotation(cmd),
		PipelineFlag:       cmd.Annotations[flags.AnnotationValueFromPipeline],
		Flags:              commandFlags(cmd),
	}
}

func commandFlags(cmd *cobra.Command) []Flag {
	pipeName := cmd.Annotations[flags.AnnotationValueFromPipeline]
	pipeOpts, _ := flags.GetPipeOptionsFromAnnotation(cmd)

	var out []Flag
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden || commonDynamicFlags[f.Name] {
			return
		}
		isPipe := pipeName != "" && strings.EqualFold(f.Name, pipeName)
		flag := Flag{
			Name:     f.Name,
			Type:     f.Value.Type(),
			Usage:    f.Usage,
			Pipeline: isPipe,
		}
		if isPipe && pipeOpts != nil && pipeOpts.Required {
			flag.Required = true
		}
		if set, ok := f.Annotations[completion.AnnotationValidateSet]; ok && len(set) > 0 {
			flag.ValidateSet = set
		}
		out = append(out, flag)
	})
	return out
}

// splitExamples breaks a cobra Example block into one string per
// blank-line-separated block.
func splitExamples(example string) []string {
	example = strings.TrimSpace(example)
	if example == "" {
		return nil
	}
	var out []string
	for block := range strings.SplitSeq(example, "\n\n") {
		block = strings.TrimRight(block, "\n")
		if strings.TrimSpace(block) != "" {
			out = append(out, block)
		}
	}
	return out
}

// DeriveMethod maps a CLI verb to the semantic REST method used to group
// projection parameter sets.
func DeriveMethod(verb string) string {
	switch strings.ToLower(verb) {
	case "create", "new":
		return "POST"
	case "update", "set":
		return "PUT"
	case "delete", "remove":
		return "DELETE"
	default:
		return "GET"
	}
}
