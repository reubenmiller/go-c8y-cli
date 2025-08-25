package login

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/mocksession"
	"github.com/stretchr/testify/assert"
)

func Test_SSOLoginFromSessionFile(t *testing.T) {
	cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST")
	if cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}
	cmd := command.NewMockCommand()
	False := false
	sessionName := "test-sso.json"
	tmpDir := mocksession.CreateCumulocitySessionDir(
		t,
		mocksession.WithSession(sessionName, &c8ysession.CumulocitySession{
			Host: cumulocityHostWithSSO,
			Settings: &config.CommandSettings{
				Encryption: &config.EncryptionSettings{
					Enabled: &False,
				},
			},
		}),
	)
	sessionFile := filepath.Join(tmpDir, sessionName)

	//
	// Login using SSO, the token should be stored in the session
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		fmt.Sprintf(`c8y sessions set %s --shell bash -v`, sessionName),
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			"C8Y_SESSION_HOME":                tmpDir,
			"C8Y_SETTINGS_ENCRYPTION_ENABLED": "false",
		}),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.Contains(t, outputEnv, "C8Y_TOKEN")
	assert.Contains(t, outputEnv, "C8Y_USER")
	assert.NotContains(t, outputEnv, "C8Y_PASSWORD")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("token").String())
	assert.NotEmpty(t, session.Get("version").String())

	//
	// Reuse the session should use the token instead of having to login again
	env := command.ParseShellEnv(stdout)
	env["C8Y_SESSION_HOME"] = tmpDir
	env["C8Y_SETTINGS_ENCRYPTION_ENABLED"] = "false"
	stdout, cmdErr = command.ExecuteCmdWithStandardOutput(
		cmd,
		fmt.Sprintf(`c8y sessions set %s --shell bash -v`, sessionName),
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, env),
	)
	assert.Nil(t, cmdErr)
	outputEnv = command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.Contains(t, outputEnv, "C8Y_TOKEN")
	assert.NotContains(t, outputEnv, "C8Y_PASSWORD")

	// Manually specify to ignore the token
	stdout, cmdErr = command.ExecuteCmdWithStandardOutput(
		cmd,
		fmt.Sprintf(`c8y sessions set %s --shell bash -v --clear`, sessionName),
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, env),
	)
	assert.Nil(t, cmdErr)
	outputEnv2 := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv2, "C8Y_HOST")
	assert.Contains(t, outputEnv2, "C8Y_TENANT")
	assert.Contains(t, outputEnv2, "C8Y_VERSION")
	assert.Contains(t, outputEnv2, "C8Y_TOKEN")
	assert.NotContains(t, outputEnv2, "C8Y_PASSWORD")

	assert.NotEqual(t, outputEnv["C8Y_TOKEN"], outputEnv2["C8Y_TOKEN"])
}

func Test_SSOLoginFromConsole(t *testing.T) {
	cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST")
	if cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}
	cmd := command.NewMockCommand()
	cmdtext := `c8y sessions login --provider env --shell bash -v`
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		cmdtext,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			"C8Y_HOST": cumulocityHostWithSSO,
		}),
	)
	assert.Nil(t, cmdErr)
	assert.Contains(t, stdout, "export C8Y_HOST=")
}

func Test_SSOLoginWithDiscoveryURL(t *testing.T) {
	cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST")
	if cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}

	sessionName := "test-sso.json"
	tmpDir := mocksession.CreateCumulocitySessionDir(
		t,
		mocksession.WithSession(sessionName, &c8ysession.CumulocitySession{
			Host: cumulocityHostWithSSO,
			Settings: &config.CommandSettings{
				SSO: &config.SSOSettings{
					DiscoveryURL: "https://invalid.com/example",
				},
			},
		}),
	)
	// sessionFile := filepath.Join(tmpDir, sessionName)

	cmd := command.NewMockCommand()
	cmdtext := fmt.Sprintf(`c8y sessions set --shell bash -v %s --clear`, "test-sso")
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		cmdtext,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			"C8Y_SESSION_HOME":                tmpDir,
			"C8Y_SETTINGS_ENCRYPTION_ENABLED": "false",
		}),
	)
	assert.Error(t, cmdErr)
	assert.Contains(t, stdout, "export C8Y_HOST=")
}

func Test_OAuth2InternalLoginWhenSSOIsActive(t *testing.T) {
	cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST")
	if cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}
	cmd := command.NewMockCommand()
	cmdtext := fmt.Sprintf(`c8y sessions set --shell bash -v --loginType OAUTH2_INTERNAL %s`, "iot-lat-sso")
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		cmdtext,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			"C8Y_HOST": cumulocityHostWithSSO,
		}),
	)
	assert.Nil(t, cmdErr)
	assert.Contains(t, stdout, "export C8Y_HOST=")
}

func Test_SetSessionUsingOAuth2InternalWhenSSOIsActive(t *testing.T) {
	if cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST"); cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}
	cmd := command.NewMockCommand()
	cmdtext := fmt.Sprintf(`c8y sessions set --shell bash -v --loginType OAUTH2_INTERNAL %s`, "iot.latest.stage.c8y.io-reuben.d.miller@gmail.com.json")
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		cmdtext,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
	)
	assert.Nil(t, cmdErr)
	assert.Contains(t, stdout, "export C8Y_HOST=")
}

func Test_SetSessionClearSSOToken(t *testing.T) {
	if cumulocityHostWithSSO := os.Getenv("SSO_C8Y_HOST"); cumulocityHostWithSSO == "" {
		t.Skipf("SSO_C8Y_HOST env variable is not set")
	}
	cmd := command.NewMockCommand()
	cmdtext := fmt.Sprintf(`c8y sessions set --shell bash -v %s`, "iot-lat-sso")
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		cmd,
		cmdtext,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
	)
	assert.Nil(t, cmdErr)
	assert.Contains(t, stdout, "export C8Y_HOST=")
}
