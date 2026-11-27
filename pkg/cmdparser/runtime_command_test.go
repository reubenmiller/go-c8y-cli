package cmdparser

import (
	"errors"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
)

// TestMissingPipelineParameter checks the error of a required pipeline
// parameter without a value, as v2.59.0 reported it: a path parameter is an
// unexpected error (exit code 100, and the unset variable of the path for a
// reference which is looked up by name), any other parameter a user error
// (exit code 101).
func TestMissingPipelineParameter(t *testing.T) {
	cmd := &RuntimeCmd{options: &CmdOptions{Spec: models.Command{
		PathParameters: []models.Parameter{
			{Name: "device", Type: "device[]"},
			{Name: "id", Type: "id[]"},
		},
		QueryParameters: []models.Parameter{
			{Name: "required_string", Type: "string"},
		},
	}}}

	exitCode := func(err error) cmderrors.ExitCode {
		var ce cmderrors.CommandError
		if errors.As(err, &ce) {
			return ce.ExitCode
		}
		return cmderrors.ExitError
	}

	cases := []struct {
		name     string
		wantText string
		wantExit cmderrors.ExitCode
	}{
		{name: "device", wantText: "missing required parameters. [device]", wantExit: cmderrors.ExitError},
		{name: "id", wantText: "Missing required parameter. id", wantExit: cmderrors.ExitError},
		{name: "required_string", wantText: "commandError: Missing required parameter. required_string", wantExit: cmderrors.ExitUserError},
	}
	for _, tc := range cases {
		err := cmd.missingPipelineParameter(&flags.ParameterError{Name: tc.name, Err: flags.ErrParameterMissing})
		text := err.Error()
		if ce, ok := err.(cmderrors.CommandError); ok {
			text = ce.Error()
		}
		if tc.wantText != "" && text != tc.wantText && text != "commandError: "+tc.wantText {
			t.Errorf("%s: text = %q, want %q", tc.name, text, tc.wantText)
		}
		if got := exitCode(err); got != tc.wantExit {
			t.Errorf("%s: exit code = %d, want %d", tc.name, got, tc.wantExit)
		}
	}
}

func TestIsReferenceLookupType(t *testing.T) {
	for _, paramType := range []string{"device[]", "devicegroup[]", "user[]", "application", "hostedapplication", "role[]"} {
		if !isReferenceLookupType(paramType) {
			t.Errorf("%s is looked up by name", paramType)
		}
	}
	for _, paramType := range []string{"string", "string[]", "id[]", "tenant", "integer", "source"} {
		if isReferenceLookupType(paramType) {
			t.Errorf("%s is not looked up by name", paramType)
		}
	}
}
