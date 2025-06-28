package noaccept

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
)

func Test_NoAccept(t *testing.T) {
	stdout, _, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		`c8y inventory create -n --template "{name: 'ci_' + _.Hex(32)}" --noAccept`,
		command.WithStdinTTY(false),
	)
	assert.Nil(t, err)
	assert.Empty(t, stdout)
}
