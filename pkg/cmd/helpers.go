package cmd

import (
	"github.com/spf13/cobra"
)

type cmder interface {
	getCommand() *cobra.Command
}
