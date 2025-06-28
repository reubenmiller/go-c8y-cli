package pipe

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_PipingNamesToCommandExpectingIds(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y events get --dry`,
		command.WithStdIn("pipeNameDoesNotExist1\npipeNameDoesNotExist2"),
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_PipingWithoutLookup(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y inventory get --dry`,
		command.WithStdIn("1111\n2222"),
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_PipingWithLookup(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y devices get --dry`,
		command.WithStdIn("agent01\ndevice01"),
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_PipingWithObjectPipe(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y devices get --dry`,
		command.WithStdIn(`{"id": "1"}`+"\n"+`{"id": "2"}`),
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
