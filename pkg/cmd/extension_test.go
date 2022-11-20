package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

func Test_ExtensionInstall(t *testing.T) {
	cmd := setupTest()
	cmdtext := `
		extension install reubenmiller/c8y-devmgmt
	`
	err := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	assert.OK(t, err)
}

func Test_ExtensionList(t *testing.T) {
	cmd := setupTest()
	cmdtext := `
		extension list
	`
	err := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	assert.OK(t, err)
}

func Test_ExtensionDelete(t *testing.T) {
	cmd := setupTest()
	cmdtext := `
		extension delete go-c8y-cli-addons
	`
	err := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	assert.OK(t, err)
}

func Test_ExtensionUpgrade(t *testing.T) {
	cmd := setupTest()
	cmdtext := heredoc.Doc(`
		extension upgrade reubenmiller/c8y-devmgmt
	`)
	err := ExecuteCmd(cmd, cmdtext)
	assert.OK(t, err)
}

func Test_ExtensionGetViews(t *testing.T) {
	cmd := setupTest()
	cmdtext := heredoc.Doc(`
		devices list --view device/agent
	`)
	err := ExecuteCmd(cmd, cmdtext)
	assert.OK(t, err)
}

func Test_ExtensionInstallFromURL(t *testing.T) {
	cmd := setupTest()
	cmdtext := `
		extension install https://github.com/reubenmiller/iot-project-c8y-cli
	`
	err := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	assert.OK(t, err)
}

func Test_Debugging(t *testing.T) {
	cmd := setupTest()
	os.Setenv("C8Y_SETTINGS_EXTENSIONS_DATADIR", "$HOME/.cumulocity/global_extensions")
	cmdtext := `
		extension upgrade devmgmt --force
	`
	// __complete template execute --template c8y-simu*::*test
	err := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	assert.OK(t, err)
}
