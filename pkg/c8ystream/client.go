package c8ystream

import (
	"context"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
	"github.com/spf13/cobra"
)

// WithProcessingMode applies the request processing mode to the context. The
// v2 core reads it from the context and sets the X-Cumulocity-Processing-Mode
// header on every request it governs. The explicit --processingMode flag wins;
// otherwise the session/config default (settings.defaults.processingMode) is
// applied so a session can pin a mode for all its commands.
func WithProcessingMode(ctx context.Context, cmd *cobra.Command, cfg *config.Config) context.Context {
	mode := ""
	if v, err := cmd.Flags().GetString(flags.FlagProcessingModeName); err == nil && v != "" {
		mode = strings.ToUpper(v)
	} else if cfg != nil {
		mode = cfg.GetProcessingMode()
	}
	if mode == "" {
		return ctx
	}
	return ctxhelpers.WithProcessingMode(ctx, types.ProcessingMode(mode))
}
