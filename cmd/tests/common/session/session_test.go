package session

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_SetUsernamePassword(t *testing.T) {
	command.RequireTenant(t)
	tenant := "t12345"
	username := "example"
	password := "dummy"
	stdout, _, err := command.ExecuteCmdWithOutputs(
		command.NewMockCommand(),
		fmt.Sprintf(
			`c8y deviceregistration getCredentials -n --id customdevice --sessionUsername "%s" --sessionPassword "%s" --dry --dryFormat json`,
			tenant+"/"+username,
			password,
		),
		command.WithStdinTTY(false),
		command.WithSensitiveLogging(t, false),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)

	response := gjson.Parse(stdout)
	assert.Equal(t, response.Get("headers.Authorization").String(), "Basic "+base64.StdEncoding.EncodeToString([]byte(tenant+"/"+username+":"+password)))
}
