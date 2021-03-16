package cmdutil

import (
	"github.com/spf13/cobra"
)

func DisableAuthCheck(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations["skipAuthCheck"] = "true"
	return cmd
}

func IsAuthCheckEnabled(cmd *cobra.Command) bool {
	if !cmd.Runnable() {
		return false
	}
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipAuthCheck"] == "true" {
			return false
		}
	}

	return true
}
