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
	// cmdArgs := "inventoryReferences assignDeviceToGroup --group=1234 --newChildDevice=testdevice_hqcr0itez3 --dry"
	cmdArgs := "users create  --dry --userName testme"
	cmd.SetArgs(strings.Split(cmdArgs, " "))
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func Test_DataFlag(t *testing.T) {
	cmd := setupTest()
	b := bytes.NewBufferString("")
	errBuffer := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetOutput(errBuffer)
	cmdArgs := "inventory create --name \"testMo\" --type \"mo_value\" --data value=1 --dry"
	cmd.SetArgs(strings.Split(cmdArgs, " "))
	cmdErr := cmd.Execute()
	assert.True(t, cmdErr != nil)

	outE := readOutput(t, errBuffer)
	assert.True(t, outE != "")

	out := readOutput(t, b)
	assert.True(t, out != "")
}

func splitCmd(line string) []string {
	return strings.Split(line, " ")
	// r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	// return r.FindAllString(line, -1)
}

func ExecuteCmd(cmd *c8yCmd, cmdArgs interface{}) error {
	switch v := cmdArgs.(type) {
	case string:
		cmd.SetArgs(splitCmd(v))

	case []string:
		cmd.SetArgs(v)
	}
	return cmd.Execute()
}

func Test_DeviceLookup(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "device get --id testdevice_1me4xsy9vd -v")
	assert.True(t, cmdErr != nil)
}

func Test_EmptyExpand(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "inventory list")
	assert.True(t, cmdErr != nil)
}

func Test_DeviceFetcher(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "devices update --id=testdevice_1me4xsy9vd --dry")
	assert.True(t, cmdErr != nil)
}

func Test_DeviceFetcherWithCollection(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "events list --type value")
	assert.OK(t, cmdErr)
}

func Test_BodyValidate(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, `events create --device 1234 --template={textd:"custom_hello"} --type value --dry`)
	assert.OK(t, cmdErr)
}

func Test_InventoryReferences(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, `inventoryReferences createChildAddition --newChild=87464 --id=87608 --pretty=false --dry`)
	assert.OK(t, cmdErr)
}

func Test_ChildInventoryReferences(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, `inventoryReferences assignDeviceToGroup --group=testdevice_1me4xsy9vd --newChildDevice testdevice_6dyojzxbvf --dry`)
	assert.OK(t, cmdErr)
}

func Test_ChildInventoryReferencesWithPipelineInput(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString("testdevice_6dyojzxbvf\ntestdevice_7ewmxq0a94\n\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `inventoryReferences assignDeviceToGroup --group=testgroup_yup6kr9sjg --dry`)
	assert.OK(t, cmdErr)
}

// Pipe options

func Test_PipingWithoutLookup(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString("1234\n222\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `inventory get --dry`)
	assert.OK(t, cmdErr)
}

func Test_PipingWithLookup(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString("testdevice_7ewmxq0a94\ntestdevice_6dyojzxbvf\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `devices get --dry`)
	assert.OK(t, cmdErr)
}

func Test_PipingWithObjectPipe(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`{"id": "87551"}` + "\n" + `{"id": "1111"}\n`)
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `devices get --dry`)
	assert.OK(t, cmdErr)
}
