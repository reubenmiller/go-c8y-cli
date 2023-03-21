package flags

import (
	"encoding/json"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	FlagDataName                  = "data"
	FlagDataTemplateName          = "template"
	FlagDataTemplateVariablesName = "templateVars"
	FlagProcessingModeName        = "processingMode"
	FlagWithTotalPages            = "withTotalPages"
	FlagWithTotalElements         = "withTotalElements"
	FlagPageSize                  = "pageSize"
	FlagCurrentPage               = "currentPage"
	FlagNullInput                 = "nullInput"
	FlagAllowEmptyPipe            = "allowEmptyPipe"
	FlagReadFromPipeText          = "-"
	FlagReadFromPipeJSON          = "-."
)
const (
	AnnotationValuePipelineAlias      = "pipelineAliases"
	AnnotationValueFromPipeline       = "valueFromPipeline"
	AnnotationValueFromPipelineData   = "valueFromPipeline.data"
	AnnotationValueCollectionProperty = "collectionProperty"
	AnnotationValueDeprecated         = "deprecatedNotice"
	AnnotationValueSemanticMethod     = "semanticMethod"
)

// Option adds flags to a given command
type Option func(*cobra.Command) *cobra.Command

// WithOptions applies given options to the command
func WithOptions(cmd *cobra.Command, opts ...Option) *cobra.Command {
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// HasValueFromPipeline checks if the given flag name supported values from pipeline
// It checks the command for a special annotation
func HasValueFromPipeline(cmd *cobra.Command, name string) bool {
	if cmd.Annotations != nil {
		if pipedArgName, ok := cmd.Annotations[AnnotationValueFromPipeline]; ok {
			return strings.EqualFold(pipedArgName, name)
		}
	}
	return false
}

// WithProcessingMode adds support for processing mode
func WithProcessingMode() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().String(FlagProcessingModeName, "", "Cumulocity processing mode")
		completion.WithOptions(
			cmd,
			completion.WithValidateSet(FlagProcessingModeName, "PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP"),
		)
		return cmd
	}
}

// WithProcessingModeValue adds the processing module value from cli arguments
func WithProcessingModeValue() GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		dst := "X-Cumulocity-Processing-Mode"

		if !cmd.Flags().Changed("processingMode") {
			return "", "", nil
		}

		value, err := cmd.Flags().GetString(FlagProcessingModeName)
		if err != nil {
			return dst, value, err
		}
		return dst, strings.ToUpper(value), err
	}
}

// WithData adds support for data input
func WithData() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().StringArrayP(FlagDataName, "d", []string{}, "static data to be applied to body. accepts json or shorthand json, i.e. --data 'value1=1,my.nested.value=100'")
		return cmd
	}
}

// WithCommonCumulocityQueryOptions adds support for common query parameter options like query, orderBy etc.
func WithCommonCumulocityQueryOptions() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().String("query", "", "Additional query filter (accepts pipeline)")
		cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
		cmd.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")

		return cmd
	}
}

// WithTemplateNoCompletion adds support for templates
func WithTemplateNoCompletion() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().String(FlagDataTemplateName, "", "Body template")
		cmd.Flags().StringArray(FlagDataTemplateVariablesName, []string{}, "Body template variables")
		return cmd
	}
}

// WithPipelineSupport adds support for pipeline to a argument
func WithPipelineSupport(name string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValueFromPipeline] = name
		return cmd
	}
}

func WithExtendedPipelineSupport(name string, property string, required bool, aliases ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValueFromPipeline] = name

		options := &PipelineOptions{
			Name:     name,
			Property: property,
			Required: required,
			Aliases:  aliases,
		}
		if required && name == "id" {
			options.IsID = true
		}
		data, err := json.Marshal(options)
		if err != nil {
			panic(err)
		}

		cmd.Annotations[AnnotationValueFromPipelineData] = string(data)
		return cmd
	}
}

func WithRuntimePipelineProperty() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		name := ""
		alias := ""
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Changed {
				switch v := f.Value.(type) {
				case pflag.SliceValue:
					values := v.GetSlice()

					// Get first value if multiple values are provided
					if len(values) > 0 {
						if values[0] == FlagReadFromPipeText {
							name = f.Name
							alias = f.Name
						} else if strings.HasPrefix(values[0], FlagReadFromPipeJSON) {
							name = f.Name
							alias = values[0][len(FlagReadFromPipeJSON):]
						}
					}

				case pflag.Value:
					if v.String() == FlagReadFromPipeText {
						name = f.Name
						alias = f.Name
					} else if strings.HasPrefix(v.String(), FlagReadFromPipeJSON) {
						name = f.Name
						alias = v.String()[len(FlagReadFromPipeJSON):]
					}
				}
			}
		})

		if name == "" {
			return cmd
		}

		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValueFromPipeline] = name

		aliases := []string{name}

		if alias != "" {
			for _, a := range strings.Split(alias, ",") {
				a = strings.TrimLeft(a, ".")
				if a != "" {
					aliases = append(aliases, a)
				}
			}
		}

		if aliasValue, ok := cmd.Annotations[AnnotationValuePipelineAlias+"."+name]; ok {
			aliases = append(aliases, strings.Split(aliasValue, ",")...)
		}

		options := &PipelineOptions{
			Name:     name,
			Property: name,
			Aliases:  aliases,
			Required: true,
			IsID:     true,
		}
		data, err := json.Marshal(options)
		if err != nil {
			panic(err)
		}

		cmd.Annotations[AnnotationValueFromPipelineData] = string(data)
		return cmd
	}
}

// WithPipelineAliases adds a list of aliases for a flag if it is selected to be sourced from the pipeline
func WithPipelineAliases(property string, aliases ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValuePipelineAlias+"."+property] = strings.Join(aliases, ",")
		return cmd
	}
}

// WithCollectionProperty adds the default property to be plucked from the raw json response so that the important information is returned by default
func WithCollectionProperty(property string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValueCollectionProperty] = property
		return cmd
	}
}

// WithDeprecationNotice marks a commands as being deprecated
func WithDeprecationNotice(message string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if cmd.Annotations == nil {
			cmd.Annotations = map[string]string{}
		}
		cmd.Annotations[AnnotationValueDeprecated] = message
		return cmd
	}
}

// GetDeprecationNoticeFromAnnotation returns the deprecated notice if present
func GetDeprecationNoticeFromAnnotation(cmd *cobra.Command) (value string) {
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		return
	}
	if v, ok := cmd.Annotations[AnnotationValueDeprecated]; ok {
		value = v
	}
	return
}

// WithSemanticMethod sets a semantic REST method which may be different to the actual REST method used
// useful to be more descriptive about how the action should behave (e.g. prompting).
func WithSemanticMethod(v string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if v != "" {
			if cmd.Annotations == nil {
				cmd.Annotations = map[string]string{}
			}
			cmd.Annotations[AnnotationValueSemanticMethod] = v
		}
		return cmd
	}
}

// GetSemanticMethodFromAnnotation returns semantic REST method related to the action from the annotations
func GetSemanticMethodFromAnnotation(cmd *cobra.Command) (value string) {
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		return
	}
	if v, ok := cmd.Annotations[AnnotationValueSemanticMethod]; ok {
		value = v
	}
	return
}

// GetPipeOptionsFromAnnotation returns the pipeline options stored in the annotations
func GetPipeOptionsFromAnnotation(cmd *cobra.Command) (options *PipelineOptions, err error) {
	options = &PipelineOptions{}
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		return
	}
	if v, ok := cmd.Annotations[AnnotationValueFromPipelineData]; ok {
		err = json.Unmarshal([]byte(v), options)
		if err != nil {
			return
		}
	}
	return
}

// GetCollectionPropertyFromAnnotation returns the collection property path used to return a subset of the json response by default
func GetCollectionPropertyFromAnnotation(cmd *cobra.Command) (value string) {
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		return
	}
	if v, ok := cmd.Annotations[AnnotationValueCollectionProperty]; ok {
		value = v
	}
	return
}

// GetStringFromAnnotation returns a string value stored in the annotations
func GetStringFromAnnotation(cmd *cobra.Command, path string) (value string) {
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		return
	}
	if v, ok := cmd.Annotations[path]; ok {
		value = v
	}
	return
}

// WithBatchOptions adds support for batch options
func WithBatchOptions(acceptInputFile bool) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		if acceptInputFile {
			cmd.Flags().String("inputFile", "", "Input file of ids to add to processed (required)")
			// cmd.MarkFlagRequired("inputFile")
		}
		cmd.Flags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")
		cmd.Flags().Int("count", 5, "Total number of objects")
		cmd.Flags().Int("startIndex", 1, "Start index value")
		cmd.Flags().Int("delay", 200, "delay in milliseconds after each request")
		cmd.Flags().Int("workers", 2, "Number of workers")
		return cmd
	}
}
