package cmdutil

import (
	"github.com/spf13/cobra"
)

// AnnotationLocalCommand marks a command which only changes local state
const AnnotationLocalCommand = "localCommand"

// MarkLocalCommand marks a command which only changes local state (aliases,
// cache, extensions, sessions, settings) and never data in Cumulocity. The
// session's create/update/delete mode protects tenant data, so it does not
// apply to such a command: e.g. switching to another session must work from a
// read-only session.
func MarkLocalCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[AnnotationLocalCommand] = "true"
	return cmd
}

// IsLocalCommand reports whether the command was marked with MarkLocalCommand
func IsLocalCommand(cmd *cobra.Command) bool {
	return cmd.Annotations[AnnotationLocalCommand] == "true"
}
