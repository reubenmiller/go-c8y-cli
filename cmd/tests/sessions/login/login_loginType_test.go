package login

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/command"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/testing/mocksession"
	"github.com/stretchr/testify/assert"
)

func createBasicSession(t *testing.T) (string, string) {
	sessionName := "test-simple.json"
	tmpDir := mocksession.CreateCumulocitySessionDir(
		t,
		mocksession.WithSession(sessionName, &c8ysession.CumulocitySession{
			Host:     os.Getenv("C8Y_HOST"),
			Username: os.Getenv("C8Y_USER"),
			Password: os.Getenv("C8Y_PASSWORD"),
		}),
	)
	return tmpDir, filepath.Join(tmpDir, sessionName)
}

func createBasicWithTokenSession(t *testing.T) (string, string) {
	session, err := mocksession.CreateSessionWithOAuth2Internal()
	assert.NoError(t, err)
	sessionName := "test-simple.json"
	tmpDir := mocksession.CreateCumulocitySessionDir(
		t,
		mocksession.WithSession(sessionName, &c8ysession.CumulocitySession{
			Host:     session.Host,
			Username: session.Username,
			Password: session.Password,
			Token:    session.Token,
		}),
	)
	return tmpDir, filepath.Join(tmpDir, sessionName)
}

func Test_LoginSwithFromTokenToBasicWithTokenPresent(t *testing.T) {
	homeDir, sessionFile := createBasicWithTokenSession(t)
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions set -v --shell bash --loginType BASIC`,
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			config.EnvSessionHome: homeDir,
		}),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.Contains(t, outputEnv, "C8Y_PASSWORD")
	assert.NotContains(t, outputEnv, "C8Y_TOKEN")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("password").String())
	assert.Empty(t, session.Get("token").String())
	assert.NotEmpty(t, session.Get("version").String())
}

func Test_LoginWithUserGivenLoginType_Default(t *testing.T) {
	homeDir, sessionFile := createBasicSession(t)
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions set -v --shell bash`,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			config.EnvSessionHome: homeDir,
		}),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.Contains(t, outputEnv, "C8Y_TOKEN")
	assert.NotContains(t, outputEnv, "C8Y_PASSWORD")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("password").String())
	assert.NotEmpty(t, session.Get("token").String(), "Token should of been updated after the login process")
	assert.NotEmpty(t, session.Get("version").String())
}

func Test_LoginWithUserGivenLoginType_OAuth2Internal(t *testing.T) {
	homeDir, sessionFile := createBasicSession(t)
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions set -v --shell bash --loginType OAUTH2_INTERNAL`,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			config.EnvSessionHome: homeDir,
		}),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.Contains(t, outputEnv, "C8Y_TOKEN")
	assert.NotContains(t, outputEnv, "C8Y_PASSWORD")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("password").String())
	assert.NotEmpty(t, session.Get("token").String(), "Token should of been updated after the login process")
	assert.NotEmpty(t, session.Get("version").String())
}

func Test_LoginWithUserGivenLoginTypeBasic(t *testing.T) {
	homeDir, sessionFile := createBasicSession(t)
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions set -v --shell bash --loginType BASIC`,
		command.WithStdIn("\n"),
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			config.EnvSessionHome: homeDir,
		}),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.NotContains(t, outputEnv, "C8Y_TOKEN")
	assert.Contains(t, outputEnv, "C8Y_PASSWORD")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.NotEmpty(t, session.Get("password").String())
	assert.Empty(t, session.Get("token").String(), "Token should not be present as the user requested to use basic auth")
	assert.NotEmpty(t, session.Get("version").String())
}

func Test_LoginWithUserGivenLoginTypeNone(t *testing.T) {
	tedgeURL := os.Getenv("TEDGE_URL")
	if tedgeURL == "" {
		t.Skip("TEDGE_URL not set")
	}

	sessionName := "test-simple.json"
	homeDir := mocksession.CreateCumulocitySessionDir(
		t,
		mocksession.WithSession(sessionName, &c8ysession.CumulocitySession{
			Host: os.Getenv("TEDGE_URL"),
		}),
	)
	sessionFile := filepath.Join(homeDir, sessionName)

	//
	// Login using SSO, the token should be stored in the session
	stdout, cmdErr := command.ExecuteCmdWithStandardOutput(
		command.NewMockCommand(),
		`c8y sessions set -v --shell bash --loginType NONE`,
		command.WithStdinTTY(true),
		command.WithStderrTTY(true),
		command.WithEnv(t, map[string]string{
			config.EnvSessionHome: homeDir,
		}),
		command.WithSessionEncryption(t, false),
	)
	assert.Nil(t, cmdErr)
	outputEnv := command.ParseShellEnv(stdout)
	assert.Contains(t, outputEnv, "C8Y_HOST")
	assert.Contains(t, outputEnv, "C8Y_TENANT")
	assert.Contains(t, outputEnv, "C8Y_VERSION")
	assert.NotContains(t, outputEnv, "C8Y_TOKEN")
	assert.NotContains(t, outputEnv, "C8Y_PASSWORD")

	// Session information should be persisted
	session := mocksession.ParseSession(sessionFile)
	assert.NotEmpty(t, session.Get("host").String())
	assert.NotEmpty(t, session.Get("tenant").String())
	assert.Empty(t, session.Get("password").String())
	assert.Empty(t, session.Get("token").String(), "Token should not be present as the user requested to use basic auth")
	assert.NotEmpty(t, session.Get("version").String())
}
