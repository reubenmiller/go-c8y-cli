package flags

import (
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
	AnnotationValueFromPipeline = "valueFromPipeline"
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
