package flags

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

const (
	FlagDataName                  = "data"
	FlagDataTemplateName          = "template"
	FlagDataTemplateVariablesName = "templateVars"
	FlagProcessingModeName        = "processingMode"
)
const (
	AnnotationValueFromPipeline       = "valueFromPipeline"
	AnnotationValueFromPipelineData   = "valueFromPipeline.data"
	AnnotationValueCollectionProperty = "collectionProperty"
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
		cmd.Flags().String(FlagProcessingModeName, "", "Processing mode")
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
		cmd.Flags().StringP(FlagDataName, "d", "", "json")
		return cmd
	}
}

// WithTemplate adds support for templates
func WithTemplate() Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().String(FlagDataTemplateName, "", "Body template")
		cmd.Flags().String(FlagDataTemplateVariablesName, "", "Body template variables")
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
