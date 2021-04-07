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

// DisableEncryptionCheck disable encryption check when reading the configuration
func DisableEncryptionCheck(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	cmd.Annotations["skipEncryptionCheck"] = "true"
	return cmd
}

// IsConfigEncryptionCheckEnabled check if the encryption in the configuration should be validated or not
func IsConfigEncryptionCheckEnabled(cmd *cobra.Command) bool {
	if !cmd.Runnable() {
		return true
	}
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skipEncryptionCheck"] == "true" {
			return false
		}
	}

	return true
}
