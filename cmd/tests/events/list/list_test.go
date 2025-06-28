package list

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_EventListWithoutDeviceIterator(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y events list -n --dateFrom=-10d --type=my_CustomType2`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_EventListWithPipedSourceID(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y events list --dry`,
		command.WithStdIn(`{"source":{"id":"1234"}}`+"\n"),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
