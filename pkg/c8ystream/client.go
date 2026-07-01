package c8ystream

import (
	"context"
	"net/url"
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

// WithExtraRequestOptions applies the CLI's global custom headers (--header) and
// custom query parameters (--customQueryParam) to the context. The v2 core reads
// them from the context and adds them to every request it governs, so they show
// up in dry-run output and are sent on real requests — matching the v1 request
// handler, which applied these globals to all outgoing requests.
func WithExtraRequestOptions(ctx context.Context, cfg *config.Config) context.Context {
	if cfg == nil {
		return ctx
	}
	if headers := parseKeyValues(cfg.GetHeader()); len(headers) > 0 {
		ctx = ctxhelpers.WithExtraHeaders(ctx, headers)
	}
	if params := parseQueryParams(cfg.GetQueryParameters()); len(params) > 0 {
		ctx = ctxhelpers.WithExtraQueryParams(ctx, params)
	}
	if cfg.IgnoreAcceptHeader() {
		ctx = ctxhelpers.WithNoAccept(ctx, true)
	}
	return ctx
}

// parseKeyValues parses "Key: Value" entries (a single setting may carry several
// comma-separated pairs, e.g. --header "A: 1, B: 2") into a map.
func parseKeyValues(entries []string) map[string]string {
	out := make(map[string]string)
	for _, entry := range entries {
		for _, pair := range strings.Split(entry, ",") {
			if k, v, ok := splitKeyValue(pair); ok {
				out[k] = v
			}
		}
	}
	return out
}

// parseQueryParams parses "key: value" / "key=value" entries into url.Values.
func parseQueryParams(entries []string) url.Values {
	out := url.Values{}
	for _, entry := range entries {
		if k, v, ok := splitKeyValue(entry); ok {
			out.Add(k, v)
		}
	}
	return out
}

// splitKeyValue splits "key: value" or "key=value" on the first delimiter,
// trimming surrounding whitespace. Returns ok=false when no key is found.
func splitKeyValue(s string) (key, value string, ok bool) {
	i := strings.IndexAny(s, ":=")
	if i < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:i])
	value = strings.TrimSpace(s[i+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}
