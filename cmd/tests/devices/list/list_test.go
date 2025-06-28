package create

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_GetAllDevices(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y devices list -n --includeAll`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
