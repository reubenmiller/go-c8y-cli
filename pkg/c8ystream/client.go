package c8ystream

import (
	"context"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/requestutil"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
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
	return apiv2.WithProcessingMode(ctx, types.ProcessingMode(mode))
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
		ctx = apiv2.WithExtraHeaders(ctx, headers)
	}
	if params := parseQueryParams(cfg.GetQueryParameters()); len(params) > 0 {
		ctx = apiv2.WithExtraQueryParams(ctx, params)
	}
	if cfg.IgnoreAcceptHeader() {
		ctx = apiv2.WithNoAccept(ctx, true)
	}
	return apiv2.WithRequestHook(ctx, requestDefaults(cfg))
}

// requestDefaults returns the request hook which gives every request the
// headers v2.59.0 sent, whichever way the SDK built it. The hook runs on the
// final request, before a dry run renders it and before it is sent:
//
//   - Accept: application/json when the request has none (some SDK calls, e.g.
//     the inventory lists and deletes, set no Accept header). --noAccept sends
//     none.
//   - not the Accept-Encoding of the SDK's HTTP client ("gzip, deflate").
//     Without it the transport asks for "gzip" and decodes the response, as
//     v2.59.0 did (and sends nothing with settings.http.compression=false). An
//     Accept-Encoding given with --header is sent, as in v2.59.0.
func requestDefaults(cfg *config.Config) func(*http.Request) error {
	return func(req *http.Request) error {
		if cfg == nil || !cfg.IgnoreAcceptHeader() {
			if req.Header.Get("Accept") == "" {
				req.Header.Set("Accept", types.MimeTypeApplicationJSON)
			}
		}
		if req.Header.Get("Accept-Encoding") == requestutil.ClientAcceptEncoding {
			req.Header.Del("Accept-Encoding")
		}
		return nil
	}
}

// WithQueryParam returns ctx with an additional query parameter for the
// requests made under it, merging with (not replacing) any globally configured
// --customQueryParam values. Use it for legacy CLI parameters that a typed SDK
// options struct deliberately does not expose (e.g. measurements list
// --fragmentType/--unit). An empty value is a no-op.
func WithQueryParam(ctx context.Context, key, value string) context.Context {
	if value == "" {
		return ctx
	}
	merged := url.Values{}
	for k, vs := range apiv2.ExtraQueryParams(ctx) {
		merged[k] = append([]string(nil), vs...)
	}
	merged.Add(key, value)
	return apiv2.WithExtraQueryParams(ctx, merged)
}

// WithHeaderParam returns ctx with an additional request header for the
// requests made under it, merging with (not replacing) any globally configured
// --header values. Use it for legacy CLI parameters that map to a header the
// typed SDK options deliberately do not expose (e.g. measurements list --unit →
// X-Cumulocity-System-Of-Units). An empty value is a no-op.
func WithHeaderParam(ctx context.Context, key, value string) context.Context {
	if value == "" {
		return ctx
	}
	merged := map[string]string{}
	maps.Copy(merged, apiv2.ExtraHeaders(ctx))
	merged[key] = value
	return apiv2.WithExtraHeaders(ctx, merged)
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
