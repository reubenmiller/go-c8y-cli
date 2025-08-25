package login

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/mocksession"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_LoginFromCommand(t *testing.T) {
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions login -n --from-cmd "c8y sessions login --from-env --output-format json --no-banner" --output-format json`,
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, mocksession.GetSessionEnvironmentVariables()),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	session := gjson.Parse(stdout)

	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("username").String())
	assert.NotEmpty(t, session.Get("token").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("version").String())
	assert.False(t, session.Get("password").Exists())
}
