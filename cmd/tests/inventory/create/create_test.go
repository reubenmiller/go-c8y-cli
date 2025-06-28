package create

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_CreateInventoryDryRun(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y inventory create -n --name testme --dry`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_CreateInventoryDryRunWithLargeNumbers(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y inventory create -n --name testme --type ci_Test --data '{
			"type": "",
			"c8y_Kpi": {
				"max": 1.91010101E+20,
				"description": ""
			}
		}'`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}

func Test_CreateManagedObjectViaPipeline(t *testing.T) {
	stdout, stderr, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y inventory create --dry`,
		command.WithStdIn("1\n2"),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
}
