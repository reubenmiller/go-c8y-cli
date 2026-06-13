package c8ystream

import (
	"context"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
	"github.com/spf13/cobra"
)

// WithProcessingMode applies the --processingMode flag to the context. The
// v2 core reads it from the context and sets the
// X-Cumulocity-Processing-Mode header on every request it governs.
func WithProcessingMode(ctx context.Context, cmd *cobra.Command) context.Context {
	if v, err := cmd.Flags().GetString(flags.FlagProcessingModeName); err == nil && v != "" {
		return ctxhelpers.WithProcessingMode(ctx, types.ProcessingMode(strings.ToUpper(v)))
	}
	return ctx
}
