// Package configurations groups the remote access configuration subcommands
// (list/get/update/delete and the create-passthrough/webssh/vnc/telnet
// variants), all hand-written v2 c8ystream commands backed by the go-c8y v2
// RemoteAccess.Configurations service.
package configurations

import (
	"context"
	"errors"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	racfg "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/remoteaccess/remoteaccess_configurations"
	"github.com/spf13/cobra"
)

// resolveConfigurationID resolves the id or name of a configuration of a
// device. A name which matches nothing is reported as v2.59.0 did for
// get, update and delete (like any other unresolved name: "referenceByName: no
// matching items found", exit code 4), not with the text of the SDK.
func resolveConfigurationID(ctx context.Context, client *apiv2.Client, deviceID, ref string) (string, error) {
	id, err := client.RemoteAccess.Configurations.ResolveID(ctx, deviceID, c8ystream.NameOrID(ref))
	if err != nil {
		var notFound *racfg.ConfigurationNotFoundError
		if errors.As(err, &notFound) {
			return "", c8ystream.UnresolvedNameError(notFound.Name)
		}
		return "", err
	}
	return id, nil
}

type SubCmdConfigurations struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdConfigurations {
	ccmd := &SubCmdConfigurations{}

	cmd := &cobra.Command{
		Use:   "configurations",
		Short: "Manage remote access configurations",
		Long:  `Cloud Remote Access configuration management`,
	}

	// Subcommands
	cmd.AddCommand(NewListCmd(f).GetCommand())
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(NewCreatePassthroughCmd(f).GetCommand())
	cmd.AddCommand(NewCreateWebsshCmd(f).GetCommand())
	cmd.AddCommand(NewCreateVncCmd(f).GetCommand())
	cmd.AddCommand(NewCreateTelnetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
