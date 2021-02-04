package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
)

func setupTest() *c8yCmd {
	configureRootCmd()
	return rootCmd
}

func Test_ExecuteCommand(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)
	// cmd.SetArgs([]string{"inventory", "update", "--id", "1234", "--template", "/workspaces/go-c8y-cli/temp-example/device.update.jsonnet", "--dry"})
	cmd.SetArgs([]string{"inventory", "create", "--name", "testme", "--dry"})
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func Test_ExecuteCollectionCommand(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)
	// cmd.SetArgs([]string{"inventory", "update", "--id", "1234", "--template", "/workspaces/go-c8y-cli/temp-example/device.update.jsonnet", "--dry"})
	cmd.SetArgs([]string{"operations", "list", "--dry"})
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func Test_ExecuteCommandWithLargeNumber(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)

	cmd.SetArgs([]string{
		"inventory", "create",
		"--name=testMO",
		"--type=customType_ikpzw0n9ah",
		"--data",
		"{\"type\":\"\",\"c8y_Kpi\":{\"max\":1.91010101E+20,\"description\":\"\"}}"})
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func Test_ExecuteTemplateIndexCommand(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)
	cmd.SetArgs([]string{"inventory", "update", "--id", "1234,4567", "--template", "/workspaces/go-c8y-cli/temp-example/update.mo.jsonnet", "--dry"})
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func readOutput(t *testing.T, b io.Reader) string {
	out, err := ioutil.ReadAll(b)
	assert.OK(t, err)
	return string(out)
}

func Test_ExecutePathVariableCommand(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)
	cmdArgs := "inventory get --id=12345 --dry"
	cmd.SetArgs(strings.Split(cmdArgs, " "))
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}
