// Package plugins runs the commands that change the UI plugins of a hosted
// application (ui applications plugins install, delete, replace and update).
// The plugins are managed by go-c8y (UIApplicationPlugins.Update); each
// command only says which change to make.
package plugins

import (
	"context"
	"encoding/json"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/ui/applicationplugins"
	"github.com/spf13/cobra"
)

type PluginCmd struct {
	*subcommand.SubCommand

	Factory *cmdutil.Factory

	Application   string
	Remove        []string
	Add           []string
	ReplaceAll    bool
	UpdateAll     bool
	RemoveInvalid bool
}

// Run changes the plugins of each application (--application or piped): one
// update per application. --data and --template give more application fields
// to change; config.remotes entries in them are added as given.
func (c *PluginCmd) Run(cmd *cobra.Command) error {
	r, err := c8ystream.NewRunner(cmd, c.Factory)
	if err != nil {
		return err
	}
	// v2.59.0 sent the requests of the ui applications plugins commands without a page size
	r.NoDefaultPageSize()
	if err := r.InputFlag("application"); err != nil {
		return err
	}
	if err := r.Body(
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(c.Factory),
		flags.WithTemplateVariablesValue(),
	); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		// v2.59.0 looked up the name of a hosted application (the ones of the
		// tenant first) before it changed it
		app, err := c8ystream.HostedApplicationRef(in.ResolveContext(), client, c.Factory.GetTenant(), in.String("application"))
		if err != nil {
			return nil, err
		}
		opt := applicationplugins.UpdateOptions{
			Install:       c.Add,
			Remove:        c.Remove,
			RemoveAll:     c.ReplaceAll,
			UpdateAll:     c.UpdateAll,
			RemoveInvalid: c.RemoveInvalid,
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		if hasFields(body) {
			opt.Body = body
		}
		return func(ctx context.Context) any {
			return client.UIApplicationPlugins.Update(ctx, app, opt)
		}, nil
	})
}

// hasFields reports whether body is a JSON object with at least one field.
func hasFields(body json.RawMessage) bool {
	var fields map[string]any
	return json.Unmarshal(body, &fields) == nil && len(fields) > 0
}
