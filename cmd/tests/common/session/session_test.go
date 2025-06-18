package session

import (
	"fmt"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_SetUsernamePassword(t *testing.T) {
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
	assert.Equal(t, response.Get("headers.Authorization").String(), c8y.NewBasicAuthString(tenant, username, password))
}
