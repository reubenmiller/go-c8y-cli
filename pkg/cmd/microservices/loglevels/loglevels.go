// Package loglevels wires the `c8y microservices loglevels` subgroup. Its
// list/get/set/delete subcommands are hand-written v2 c8ystream commands backed
// by the go-c8y v2 Microservices.Loggers service (the Spring Boot actuator
// /service/{name}/loggers endpoints), so this group command is hand-written
// rather than generated.
package loglevels

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/loglevels/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/loglevels/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/loglevels/list"
	cmdSet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/loglevels/set"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdLoglevels struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdLoglevels {
	ccmd := &SubCmdLoglevels{}

	cmd := &cobra.Command{
		Use:   "loglevels",
		Short: "Cumulocity microservice log levels",
		Long: `Manage log levels of microservices.
Loggers define the log levels based on the qualified name of the Java class.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)
`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdSet.NewSetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
