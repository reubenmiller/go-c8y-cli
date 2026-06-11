package powershell

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/codegen"
)

// The Pester test templates (kept byte identical to the originals under
// scripts/build-powershell/templates, including their BOM which is stripped
// on load the way Get-Content -Raw does).
var (
	//go:embed templates/test.template.ps1
	testTemplateRaw string

	//go:embed templates/testcase.template.ps1
	testCaseTemplateRaw string

	//go:embed templates/testcase.emptyresponse.template.ps1
	testCaseEmptyResponseTemplateRaw string
)

func loadTemplate(raw string) string {
	return strings.TrimPrefix(raw, "\uFEFF")
}

var (
	reItStatement    = regexp.MustCompile(`\bIt "`)
	reRandomDevice   = regexp.MustCompile(`\{\{\s*randomdevice\s*\}\}`)
	reRandomDeviceQ  = regexp.MustCompile(`"?\{\{\s*randomdevice\s*\}\}"?`)
	reRandomAgent    = regexp.MustCompile(`\{\{\s*randomagent\s*\}\}`)
	reRandomAgentQ   = regexp.MustCompile(`"?\{\{\s*randomagent\s*\}\}"?`)
	reNewAlarm       = regexp.MustCompile(`\{\{\s*NewAlarm\s*\}\}`)
	reNewAlarmQ      = regexp.MustCompile(`"?\{\{\s*NewAlarm\s*\}\}"?`)
	reNewOperation   = regexp.MustCompile(`\{\{\s*NewOperation\s*\}\}`)
	reNewOperationQ  = regexp.MustCompile(`"?\{\{\s*NewOperation\s*\}\}"?`)
	reNewEvent       = regexp.MustCompile(`\{\{\s*NewEvent\s*\}\}`)
	reNewEventQ      = regexp.MustCompile(`"?\{\{\s*NewEvent\s*\}\}"?`)
	reVarCommand     = regexp.MustCompile(`\{\{\s*Command\s*\}\}`)
	reVarDescription = regexp.MustCompile(`\{\{\s*Description\s*\}\}`)
	reVarCmdletName  = regexp.MustCompile(`\{\{\s*CmdletName\s*\}\}`)
	reVarTestCases   = regexp.MustCompile(`\{\{\s*TestCases\s*\}\}`)
	reVarBeforeEach  = regexp.MustCompile(`\{\{\s*BeforeEach\s*\}\}`)
	reVarAfterEach   = regexp.MustCompile(`\{\{\s*AfterEach\s*\}\}`)
)

// testCase carries one powershell example of a command specification.
type testCase struct {
	Command     string
	Description string
	BeforeEach  []string
	AfterEach   []string
}

// renderTest ports New-C8yApiPowershellTest.ps1 together with the test case
// collection of New-C8yApiPowershellCommand.ps1. It returns the content of
// the Tests/<CmdletName>.auto.Tests.ps1 file (without BOM and trailing
// newline) and whether a test file applies to the command at all.
func renderTest(command *codegen.Value, cmdletName string) (string, bool) {
	examples := command.Get("examples").Get("powershell")
	if !examples.Truthy() {
		return "", false
	}

	var testCases []*testCase
	for _, example := range examples.Items() {
		if !example.Get("command").Truthy() {
			continue
		}
		tc := &testCase{
			Command:     example.Get("command").Str(),
			Description: strings.ReplaceAll(example.Get("description").Str(), "\"", "'"),
		}
		for _, statement := range example.Get("beforeEach").Items() {
			tc.BeforeEach = append(tc.BeforeEach, statement.Str())
		}
		for _, statement := range example.Get("afterEach").Items() {
			tc.AfterEach = append(tc.AfterEach, statement.Str())
		}
		testCases = append(testCases, tc)
	}
	if len(testCases) == 0 {
		return "", false
	}

	// Adjust test case template depending if a response is expected or not
	caseTemplate := loadTemplate(testCaseTemplateRaw)
	if strings.TrimSpace(command.Get("accept").Str()) == "" {
		caseTemplate = loadTemplate(testCaseEmptyResponseTemplateRaw)
	}

	// $iExample holds the last example after the collection loop, so the
	// skipTest flag of the last example applies to all test cases.
	items := examples.Items()
	skipTest := items[len(items)-1].Get("skipTest").StrEquals("true")

	beforeBlock := &strings.Builder{}
	afterBlock := &strings.Builder{}

	rendered := make([]string, 0, len(testCases))
	for _, tc := range testCases {
		caseText := caseTemplate

		// Skip Test
		if skipTest {
			caseText = reItStatement.ReplaceAllLiteralString(caseText, `It -Skip "`)
		}

		// Add any explicit before blocks
		for _, statement := range tc.BeforeEach {
			if !strings.Contains(beforeBlock.String(), statement) {
				beforeBlock.WriteString("        " + statement + "\n")
			}
		}

		// Add any explicit after blocks
		for _, statement := range tc.AfterEach {
			if !strings.Contains(afterBlock.String(), statement) {
				afterBlock.WriteString("        " + statement + "\n")
			}
		}

		// Replace any random variables
		if reRandomDevice.MatchString(tc.Command) {
			beforeStatement := `$TestDevice = PSc8y\New-TestDevice`
			if !strings.Contains(beforeBlock.String(), beforeStatement) {
				beforeBlock.WriteString("        " + beforeStatement + "\n")
			}

			tc.Command = reRandomDeviceQ.ReplaceAllLiteralString(tc.Command, "$TestDevice.id")

			afterStatement := `if ($TestDevice.id) {`
			if !strings.Contains(afterBlock.String(), afterStatement) {
				afterBlock.WriteString("        " + afterStatement + "\n")
				afterBlock.WriteString("            PSc8y\\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue\n")
				afterBlock.WriteString("        }\n")
			}
		}

		if reRandomAgent.MatchString(tc.Command) {
			beforeBlock.WriteString("        $TestAgent = PSc8y\\New-TestAgent\n")

			tc.Command = reRandomAgentQ.ReplaceAllLiteralString(tc.Command, "$TestAgent.id")

			afterBlock.WriteString("        if ($TestAgent.id) {\n")
			afterBlock.WriteString("            PSc8y\\Remove-ManagedObject -Id $TestAgent.id -ErrorAction SilentlyContinue\n")
			afterBlock.WriteString("        }\n")
		}

		if reNewAlarm.MatchString(tc.Command) {
			beforeBlock.WriteString("        $TestAlarm = PSc8y\\New-TestAlarm\n")

			tc.Command = reNewAlarmQ.ReplaceAllLiteralString(tc.Command, "$TestAlarm.id")

			afterBlock.WriteString("        if ($TestAlarm.source.id) {\n")
			afterBlock.WriteString("            PSc8y\\Remove-ManagedObject -Id $TestAlarm.source.id -ErrorAction SilentlyContinue\n")
			afterBlock.WriteString("        }\n")
		}

		if reNewOperation.MatchString(tc.Command) {
			beforeBlock.WriteString("        $TestOperation = PSc8y\\New-TestOperation\n")

			tc.Command = reNewOperationQ.ReplaceAllLiteralString(tc.Command, "$TestOperation.id")

			afterBlock.WriteString("        if ($TestOperation.deviceId) {\n")
			afterBlock.WriteString("            PSc8y\\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue\n")
			afterBlock.WriteString("        }\n")
		}

		if reNewEvent.MatchString(tc.Command) {
			beforeBlock.WriteString("        $TestEvent = PSc8y\\New-TestEvent\n")

			tc.Command = reNewEventQ.ReplaceAllLiteralString(tc.Command, "$TestEvent.id")

			afterBlock.WriteString("        if ($TestEvent.source.id) {\n")
			afterBlock.WriteString("            PSc8y\\Remove-ManagedObject -Id $TestEvent.source.id -ErrorAction SilentlyContinue\n")
			afterBlock.WriteString("        }\n")
		}

		// Create test case (the hashtable enumeration order of the original
		// script: Command, Description)
		caseText = dotnetReplace(reVarCommand, caseText, tc.Command)
		caseText = dotnetReplace(reVarDescription, caseText, tc.Description)
		rendered = append(rendered, caseText)
	}

	out := loadTemplate(testTemplateRaw)
	out = dotnetReplace(reVarCmdletName, out, cmdletName)
	out = dotnetReplace(reVarTestCases, out, strings.Join(rendered, "\n"))
	out = dotnetReplace(reVarBeforeEach, out, beforeBlock.String())
	out = dotnetReplace(reVarAfterEach, out, afterBlock.String())

	return out, true
}
