// Package factory builds the assert commands (inventory, devices, alarms,
// events, measurements, operations and tenants assert ...). They share their
// flags and behavior; each command only says what it checks.
package factory

import (
	"context"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/spf13/cobra"
)

// Target is the flag that names what is checked.
type Target int

const (
	// ManagedObject is checked by --id.
	ManagedObject Target = iota
	// Device is checked by --device.
	Device
	// Tenant is checked by --tenant.
	Tenant
)

// Check builds the state check of one item. ref is the value of the target
// flag (a managed object reference or a tenant id).
type Check func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check

// Assert describes one assert command.
type Assert struct {
	Target Target
	// PassState outputs the current state of the object when the assertion
	// holds. Otherwise the item's input is passed on untouched.
	PassState bool
	Check     Check
}

// New adds the shared assert flags to cmd and runs it.
//
// When the assertion holds, the item is passed on, so the passing items can
// be piped to another command. A failed assertion outputs nothing and is not
// an error unless --strict is set. Errors exit with code 112.
func New(cmd *cobra.Command, f *cmdutil.Factory, a Assert) *cobra.Command {
	flag := "id"
	switch a.Target {
	case Device:
		flag = "device"
		cmd.Flags().StringSlice(flag, []string{""}, "Device id or name (accepts pipeline)")
		completion.WithOptions(cmd, completion.WithDevice(flag, f.ClientV2))
		flags.WithOptions(cmd, flags.WithExtendedPipelineSupport(flag, flag, false, "deviceId", "source.id", "managedObject.id", "id"))
	case Tenant:
		flag = "tenant"
		cmd.Flags().StringSlice(flag, []string{}, "Tenant id (accepts pipeline)")
		completion.WithOptions(cmd, completion.WithTenantID(flag, f.ClientV2))
		flags.WithOptions(cmd, flags.WithExtendedPipelineSupport(flag, flag, false, "tenantId", "id", "owner.tenant.id", "tenant"))
	default:
		cmd.Flags().StringSlice(flag, []string{""}, "Inventory id (required) (accepts pipeline)")
		flags.WithOptions(cmd, flags.WithExtendedPipelineSupport(flag, flag, true, "deviceId", "source.id", "managedObject.id", "id"))
	}
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	cmd.Flags().String("interval", "5s", "Interval to check on the status, i.e. 10s or 1min")
	cmd.Flags().Int64("attempts", -1, "Number of attempts before giving up per id (-1 = unlimited)")
	cmd.Flags().Bool("strict", false, "Strict mode, fail if no match is found")
	cmd.SilenceUsage = true

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts, err := WaitOptions(cmd)
		if err != nil {
			return err
		}
		strict, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}
		r, err := c8ystream.NewRunner(cmd, f)
		if err != nil {
			return err
		}
		if err := r.InputFlag(flag); err != nil {
			return err
		}
		client, err := r.Client()
		if err != nil {
			return err
		}
		r.Assert(strict)
		return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
			ref := in.String(flag)
			if a.Target != Tenant {
				ref = c8ystream.NameOrID(ref)
			}
			check := a.Check(client, in, ref)
			input := in.Input()
			if len(input) == 0 {
				input = []byte(in.String(flag))
			}
			return func(ctx context.Context) any {
				doc, err := c8ywait.Until(ctx, opts, check)
				if err != nil {
					return c8ywait.Seq(jsondoc.Empty(), err)
				}
				if a.PassState || len(input) == 0 {
					return c8ywait.Seq(doc, nil)
				}
				return c8ywait.Seq(jsondoc.New(input), nil)
			}, nil
		})
	}
	return cmd
}

// WaitOptions reads the --duration, --interval and --attempts flags.
func WaitOptions(cmd *cobra.Command) (c8ywait.Options, error) {
	var opts c8ywait.Options
	var err error
	if opts.Timeout, err = flags.GetDurationFlag(cmd, "duration", true, time.Second); err != nil {
		return opts, err
	}
	if cmd.Flags().Lookup("interval") != nil {
		if opts.Interval, err = flags.GetDurationFlag(cmd, "interval", true, time.Second); err != nil {
			return opts, err
		}
	}
	if cmd.Flags().Lookup("attempts") != nil {
		if opts.Attempts, err = cmd.Flags().GetInt64("attempts"); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// RunWait runs a wait command (operations wait, inventory wait): each item
// named by flag is checked until it reaches the wanted state, and its last
// state is output, also when the wait fails.
func RunWait(cmd *cobra.Command, f *cmdutil.Factory, flag string, check Check) error {
	opts, err := WaitOptions(cmd)
	if err != nil {
		return err
	}
	r, err := c8ystream.NewRunner(cmd, f)
	if err != nil {
		return err
	}
	if err := r.InputFlag(flag); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		state := check(client, in, in.String(flag))
		return func(ctx context.Context) any {
			doc, err := c8ywait.Until(ctx, opts, state)
			// A failed request ends the wait with a command error (exit code
			// 100), not a server error, as in v1.
			return c8ywait.Seq(doc, c8ystream.WaitRequestError(err))
		}, nil
	})
}
