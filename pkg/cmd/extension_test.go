package cmd

import (
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

func Test_ExtensionUpgrade(t *testing.T) {
	cmd := setupTest()
	cmdtext := heredoc.Doc(`
		extension upgrade reubenmiller/c8y-devmgmt
	`)
	err := ExecuteCmd(cmd, cmdtext)
	assert.OK(t, err)
}
