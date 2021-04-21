package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/root"
)

func setupTest() *root.CmdRoot {

	rootCmd, err := Initialize()
	if err != nil {
		panic(err)
	}
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
	if args, err := shlex.Split(line); err == nil {
		return args
	}
	return strings.Split(line, " ")
	// r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	// return r.FindAllString(line, -1)
}

func ExecuteCmd(cmd *root.CmdRoot, cmdArgs interface{}) error {
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
	cmdErr := ExecuteCmd(cmd, `events create --device 1234 --template={text:"custom_hello"} --type value --dry`)
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

func Test_LookupQueryParameterByReference(t *testing.T) {
	cmd := setupTest()
	// stdin := bytes.NewBufferString(`{"id": "87551"}` + "\n" + `{"id": "1111"}\n`)
	// cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `operations list --device=testdevice_1me4xsy9vd`)
	assert.OK(t, cmdErr)
}

func Test_PageSizeParameter(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, `devices list --select id,name`)
	assert.OK(t, cmdErr)
}

func Test_PipeDeviceNameToQueryParameter(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`testdevice_7ewmxq0a94` + "\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `operations list --dry`)
	assert.OK(t, cmdErr)
}

func Test_UpdateEventBinary(t *testing.T) {
	cmd := setupTest()

	// setup := c8ytestutils.NewTestSetup()
	// setup.NewRandomTestDevice()

	stdin := bytes.NewBufferString(`testdevice_7ewmxq0a94` + "\n")
	cmd.SetIn(stdin)

	f, err := ioutil.TempFile(os.TempDir(), "eventBinary")
	assert.OK(t, err)
	_, err = f.WriteString("äüß1234dfÖ")
	assert.OK(t, err)
	f.Close()
	defer os.Remove(f.Name())

	cmdErr := ExecuteCmd(cmd, fmt.Sprintf("events updateBinary --id 88578 --file %s --dry", f.Name()))
	assert.OK(t, cmdErr)
}

func Test_GetCurrentUserInventoryRole(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString(`{"id":1}` + "\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "users getCurrentUserInventoryRole")
	assert.OK(t, cmdErr)
}

func Test_EventListWithoutDeviceIterator(t *testing.T) {
	cmd := setupTest()

	// stdin := bytes.NewBufferString(`1` + "\n")
	// cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "events list --dateFrom=-10d --type=my_CustomType2")
	assert.OK(t, cmdErr)
}

func Test_PipeSourceId(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`{"source":{"id":"1234"}}` + "\n")
	cmd.SetIn(stdin)
	cmdErr := ExecuteCmd(cmd, `events list --dry`)
	assert.OK(t, cmdErr)
}

func Test_PipingWithLookupNonExistant(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString("pipeNameDoesNotExist1\npipeNameDoesNotExist2")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "events list --dry")
	assert.OK(t, cmdErr)
}

func Test_NilQueryParameters(t *testing.T) {
	cmd := setupTest()

	cmdErr := ExecuteCmd(cmd, "auditRecords list --dry")
	assert.OK(t, cmdErr)
}

func Test_NilManagedObject(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "inventory create --dry")
	assert.OK(t, cmdErr)
}

func Test_PipingNamesToCommandExpectingIds(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString("pipeNameDoesNotExist1\npipeNameDoesNotExist2")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "events get --dry")
	assert.OK(t, cmdErr)
}

func Test_NoAccept(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "inventory create --name test01 --noAccept")
	assert.OK(t, cmdErr)
}

func Test_GetAllDevices(t *testing.T) {
	cmd := setupTest()
	cmdErr := ExecuteCmd(cmd, "devices list --includeAll")
	assert.OK(t, cmdErr)
}

func Test_CreateManagedObjectViaPipeline(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString("1\n2")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "inventory create --dry")
	assert.OK(t, cmdErr)
}

func Test_CreateDeviceViaPipeline(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString("1\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "devices create --template {type:input.index} --dry")
	assert.OK(t, cmdErr)
}

func Test_CreateManagedObjectWithoutInput(t *testing.T) {
	cmd := setupTest()

	// cmdErr := ExecuteCmd(cmd, fmt.Sprintf("devices list --select id,nam* --output csvheader"))
	// cmdErr := ExecuteCmd(cmd, fmt.Sprintf("applications get --id cockpit --select appId:id,tenantId:owner.**.id"))
	// stdin := bytes.NewBufferString(`{"self":"https://t1640885.latest.stage.c8y.io/user/roles/ROLE_ALARM_ADMIN"}` + "\n")
	// cmd.SetIn(stdin)
	cmdtext := `
	rest POST https://example.com/alarm/alarms --data name=test
	`
	cmdErr := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	cmdErr = CheckCommandError(cmd.Command, nil, cmdErr)
	assert.OK(t, cmdErr)
}

func Test_PipedDataToTemplate(t *testing.T) {
	cmd := setupTest()

	stdin := bytes.NewBufferString("10\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, "devices create --template {type:input.index} --dry")
	assert.OK(t, cmdErr)
}

/*
Using piped input in tempaltes
*/
func Test_PipingWithObjectPipeToTemplate(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`{"id": "87551"}` + "\n" + `{"id": "1111"}` + "\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `devices create --dry`)
	assert.OK(t, cmdErr)
}

func Test_UpdatePipingWithObjectPipeToTemplate(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`{"id": "87551"}` + "\n" + `{"id": "1111"}` + "\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `devices update --dry`)
	assert.OK(t, cmdErr)
}

func Test_PipingWithObjectPipeToTemplateWithIDs(t *testing.T) {
	cmd := setupTest()
	stdin := bytes.NewBufferString(`1111` + "\n" + `2222` + "\n")
	cmd.SetIn(stdin)

	cmdErr := ExecuteCmd(cmd, `devices update --dry`)
	assert.OK(t, cmdErr)
}

func Test_DebugCommand(t *testing.T) {
	cmd := setupTest()

	// cmdErr := ExecuteCmd(cmd, fmt.Sprintf("devices list --select id,nam* --output csvheader"))
	// cmdErr := ExecuteCmd(cmd, fmt.Sprintf("applications get --id cockpit --select appId:id,tenantId:owner.**.id"))
	stdin := bytes.NewBufferString("\n")
	cmd.SetIn(stdin)
	cmdtext := `
	
	`
	cmdErr := ExecuteCmd(cmd, strings.TrimSpace(cmdtext))
	cmdErr = CheckCommandError(cmd.Command, nil, cmdErr)
	assert.OK(t, cmdErr)
}
