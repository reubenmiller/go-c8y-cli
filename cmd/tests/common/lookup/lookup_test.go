package lookup

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_DeviceLookupByNameUsingGet(t *testing.T) {
	command.RequireTenant(t)
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y devices get -n --id agent01 -v`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	// verbose logging is allowed (e.g. the http client's info about an http host), but no warnings or errors
	assert.NotContains(t, stderr, "level=WARN")
	assert.NotContains(t, stderr, "level=ERROR")
}

func Test_DeviceLookupByNameUsingPut(t *testing.T) {
	command.RequireTenant(t)
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y devices update -n --id agent01 --dry`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
