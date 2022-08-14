package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/shlex"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flatten"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"gopkg.in/yaml.v3"
)

var logger *zap.Logger
var loggerS *zap.SugaredLogger

func init() {
	createLogger()
}

func createLogger() {
	consoleEncCfg := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleLevel := zapcore.InfoLevel
	consoleLevelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= consoleLevel
	})
	var cores []zapcore.Core
	cores = append(cores, zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncCfg),
		zapcore.Lock(zapcore.AddSync(os.Stderr)),
		consoleLevelEnabler,
	))
	core := zapcore.NewTee(cores...)
	logger = zap.New(core)
	defer func() {
		_ = logger.Sync()
	}()
	loggerS = logger.Sugar()
}

type Generator struct {
	Spec  *models.Specification
	Mocks *models.MockConfiguration
}

func NewGenerator(name string, mockConfig *models.MockConfiguration) (gen *Generator, err error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		loggerS.Fatalf("Failed to read spec file. file=%s err=%s", name, err)
		return
	}
	spec := &models.Specification{}
	err = yaml.Unmarshal(contents, spec)
	if err != nil {
		loggerS.Fatalf("Failed to marshal spec file. file=%s err=%s", name, err)
		return
	}

	gen = &Generator{
		Spec:  spec,
		Mocks: mockConfig,
	}
	return
}

func (g *Generator) CreateTests(outDir string) error {

	for _, endpoint := range g.Spec.Endpoints {
		if endpoint.Skip != nil && *endpoint.Skip {
			continue
		}
		testsuite := &models.TestSuite{
			Tests: map[string]models.TestCase{},
		}

		for i, example := range endpoint.Examples.Go {
			testcase := &models.TestCase{
				ExitCode: 0,
				Skip:     example.SkipTest,
			}

			key := CreateKey(i, g.Spec.Information.Name, &endpoint)
			testcase.Command = example.Command

			// Replace mock files
			testcase.Command = g.Mocks.ReplaceFiles(testcase.Command)

			// wrap in a shell if using pipes
			if strings.Contains(testcase.Command, "|") {
				command := g.Mocks.Replace(testcase.Command)
				testcase.Command = fmt.Sprintf("$TEST_SHELL -c '%s'", command)
			}

			loggerS.Debugf("Processing endpoint: %s", testcase.Command)
			testcase.StdOut = buildAssertions(g.Spec.Information.Name, &endpoint, i)

			testsuite.Tests[key] = *testcase
		}

		suitekey := CreateSuiteKey(g.Spec, &endpoint)

		if err := WriteTestSuite(testsuite, suitekey, outDir); err != nil {
			loggerS.Fatalf("Failed to write test suite to file. %s", err)
		}
	}

	return nil
}

func parseCommand(parentCmd string, command string, endpoint *models.EndPoint) []string {
	commands := strings.Split(command, "|")
	if len(commands) > 1 {
		currentCmd := ""
		for _, subcmd := range commands {
			if strings.Contains(subcmd, strings.Join([]string{parentCmd, endpoint.Alias.Go}, " ")) {
				currentCmd = subcmd
				break
			}
		}
		parts, _ := shlex.Split(currentCmd)
		return parts
	} else {
		parts, _ := shlex.Split(command)
		return parts
	}
}

func CreateFakeCommand(parentCmd string, endpoint *models.EndPoint) *cobra.Command {
	cmd := &cobra.Command{
		Use: endpoint.Alias.Go,
	}

	// Add global commands, don't need all of them just the ones used in the examples
	cmd.PersistentFlags().StringArray("select", nil, "Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'")
	cmd.PersistentFlags().String("output", "o", "Output format i.e. table, json, csv, csvheader")

	for _, parameter := range endpoint.GetAllParameters() {
		loggerS.Debugf("Adding parameter. name=%s", parameter.Name)
		if strings.Contains(parameter.Type, "[]") {
			cmd.Flags().StringSlice(parameter.Name, nil, "")
		} else if parameter.Type == "boolean" || parameter.Type == "optional_fragment" || parameter.Type == "booleanDefault" {
			defaultValue := parameter.Default == "true"
			cmd.Flags().Bool(parameter.Name, defaultValue, "")
		} else {
			cmd.Flags().String(parameter.Name, "", "")
		}
	}

	// Common parameters
	if endpoint.IsCollection() {
		cmd.Flags().IntP("pageSize", "p", 0, "")
	}
	if endpoint.SupportsTemplates() {
		cmd.Flags().String("template", "", "")
		cmd.Flags().String("templateVars", "", "")
	}

	cmd.Flags().String("outputFileRaw", "", "")
	cmd.Flags().String("outputFile", "", "")
	return cmd
}

func parseFakeCommand(parentCmd string, command string, endpoint *models.EndPoint) *cobra.Command {
	cmd := CreateFakeCommand(parentCmd, endpoint)
	if err := cmd.ParseFlags(parseCommand(parentCmd, command, endpoint)); err != nil {
		loggerS.Fatalf("Failed to parse command. command=%s, err=%s", command, err)
	}

	return cmd
}

func buildAssertions(parentCmd string, endpoint *models.EndPoint, exampleIdx int) (assertions *models.OutputAssertion) {
	cmd := parseFakeCommand(parentCmd, endpoint.Examples.Go[exampleIdx].Command, endpoint)

	assertions = &models.OutputAssertion{
		Contains: []string{},
		JSON: map[string]string{
			"method": endpoint.Method,
		},
	}
	if endpoint.Examples.Go[exampleIdx].AssertStdout != nil {
		for k, v := range endpoint.Examples.Go[exampleIdx].AssertStdout.JSON {
			assertions.JSON[k] = v
		}

		if len(endpoint.Examples.Go[exampleIdx].AssertStdout.Contains) > 0 {
			assertions.Contains = endpoint.Examples.Go[exampleIdx].AssertStdout.Contains
		}
	}

	expectedPath := substituteVariables(cmd, endpoint)

	if _, ok := assertions.JSON["path"]; !ok {
		if strings.Contains(expectedPath, "?") {
			i := strings.Index(expectedPath, "?")
			assertions.JSON["path"] = expectedPath[0:i]
		} else {
			assertions.JSON["path"] = expectedPath
		}
	}

	if _, ok := assertions.JSON["query"]; !ok && strings.Contains(expectedPath, "?") {
		i := strings.Index(expectedPath, "?")
		assertions.JSON["query"] = strings.ReplaceAll(expectedPath[i+1:], " ", " ")
	}

	usesCustomContains := len(assertions.Contains) > 0

	// Query parameters
	if !usesCustomContains {

		for _, parameter := range endpoint.GetQueryParameters() {
			value := getParameterValue(cmd, &parameter)
			if value != "" {
				if parameter.IsTypeDateTime() {
					// for relative dates, just if parameter was defined, ignore the value
					assertions.Contains = append(assertions.Contains, fmt.Sprintf("%s=", parameter.GetTargetProperty()))
				} else {
					switch parameter.Type {
					case "[]string":
						for _, v := range strings.Split(value, ",") {
							assertions.Contains = append(assertions.Contains, fmt.Sprintf("%s=%s", parameter.GetTargetProperty(), v))
						}

					default:
						assertions.Contains = append(assertions.Contains, fmt.Sprintf("%s=%s", parameter.GetTargetProperty(), value))
					}
				}
			}
		}
	}

	// Body
	if endpoint.IsBodyFormData() {
		// TODO: Apply assertions for formdata
	} else {
		for _, parameter := range endpoint.Body {
			value := getParameterValue(cmd, &parameter)
			loggerS.Debugf("Adding body property. name=%s, value=%s", parameter.Name, value)
			if value != "" {
				switch parameter.Type {
				case "attachment", "file", "fileContents":
					// TODO: Add support for checking file type
				case "datetime":
					// Note: Simplify checkings for relative time values
					assertions.Contains = append(assertions.Contains, fmt.Sprintf("\"%s\":", parameter.GetTargetProperty()))
				default:
					formatJsonAssertion(assertions.JSON, parameter.Type, "body."+parameter.GetTargetProperty(), value)
				}
			}
		}
	}

	return assertions
}

func formatJsonAssertion(jsonAssertion map[string]string, propType string, prop string, value string) {
	values := []string{}
	if strings.Contains(value, ",") && !(strings.Contains(value, "{") && strings.Contains(value, "}")) {
		values = strings.Split(value, ",")
	} else {
		values = append(values, value)
	}
	if len(values) == 1 {
		if strings.HasSuffix(prop, ".data") || strings.EqualFold(propType, "json_custom") {
			data := make(map[string]interface{})
			if err := jsonUtilities.ParseJSON(values[0], data); err != nil {
				loggerS.Fatalf("Could not parse shorthand json. %s", err)
			}

			prefix := "body."
			if !strings.HasSuffix(prop, ".data") {
				prefix = prop + "."
			}
			flatData, err := flatten.Flatten(data, prefix, flatten.DotStyle)
			if err != nil {
				loggerS.Fatalf("Could not flatten map. %s", err)
			}
			for k, v := range flatData {
				switch tv := v.(type) {
				case map[string]interface{}:
					if vjson, err := json.Marshal(tv); err == nil {
						jsonAssertion[k] = fmt.Sprintf("%s", vjson)
					}
					// if len(tv) == 0 {
					// 	jsonAssertion[k] = "{}"
					// } else {

					// }
				default:
					jsonAssertion[k] = fmt.Sprintf("%v", v)
				}
			}
		} else {
			if _, ok := jsonAssertion[prop]; !ok {
				jsonAssertion[prop] = values[0]
			}
		}
		return
	}
	for _, value := range values {
		// Apply GJSON query in the format of:
		// ..#(body.managedObject.id="12222").body.managedObject.id: "12222"
		query := fmt.Sprintf(`..#(%[1]s="%[2]s").%[1]s`, prop, value)
		jsonAssertion[query] = value
	}
}

func getParameterValue(cmd *cobra.Command, parameter *models.Parameter) (value string) {
	if !cmd.Flags().Changed(parameter.Name) {
		return
	}
	switch parameter.Type {
	case "integer":
		v, err := cmd.Flags().GetInt64(parameter.Name)
		if err == nil {
			value = fmt.Sprintf("%d", v)
		}
	// case "applications":
	// 	fallthrough
	case "string", "application":
		v, err := cmd.Flags().GetString(parameter.Name)
		if err == nil {
			value = v
		}
	case "boolean", "booleanDefault", "optional_fragment":
		v, err := cmd.Flags().GetBool(parameter.Name)
		if err == nil {
			if parameter.Value != "" {
				value = parameter.Value
			} else {
				value = fmt.Sprintf("%v", v)
			}
		}
	default:
		if strings.ContainsAny(parameter.Type, "[]") {
			v, err := cmd.Flags().GetStringSlice(parameter.Name)
			if err == nil {
				value = strings.Join(v, ",")
			}
		} else {
			v, err := cmd.Flags().GetString(parameter.Name)
			if err == nil {
				value = v
			}
		}

	}
	return
}

func substituteVariables(cmd *cobra.Command, endpoint *models.EndPoint) (out string) {

	out = endpoint.Path
	if !strings.HasPrefix(out, "/") {
		out = "/" + out
	}
	if strings.ContainsAny(out, "{}") {
		// detect variables
		iStart := -1
		variableNames := map[string]bool{}
		for i, c := range out {
			if iStart == -1 && c == '{' {
				iStart = i
			}
			if iStart != -1 && c == '}' {
				variableNames[out[iStart+1:i]] = true
				iStart = -1
			}
		}

		loggerS.Debugf("Detected variables: %v", variableNames)
		// replace variable values from fake command

		for _, parameter := range endpoint.PathParameters {
			value := getParameterValue(cmd, &parameter)
			if value != "" {

				if parameter.Type == "inventoryChildType" {
					// TODO: find better way to generate derivative values
					value = "child" + strings.ToUpper(value[0:1]) + strings.ToLower(value[1:]) + "s"
				}

				out = strings.Replace(out, "{"+parameter.GetTargetProperty()+"}", value, -1)
			}
		}

		// replace fixed environment variables
		out = strings.ReplaceAll(out, "{tenant}", "$C8Y_TENANT")
	}
	return
}

func WriteTestSuite(t *models.TestSuite, id string, outDir string) (err error) {
	outFile := path.Join(outDir, id+".yaml")
	out, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
		loggerS.Fatal(err)
	}

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	fmt.Fprintf(f, "%s", out)
	return
}

func CreateSuiteKey(spec *models.Specification, endpoint *models.EndPoint) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(spec.Information.Name), endpoint.Alias.Go)
}

func CreateKey(index int, parent string, endpoint *models.EndPoint) string {
	return fmt.Sprintf(
		"%s_%s_%s",
		parent,
		endpoint.Alias.Go,
		endpoint.Examples.Go[index].Description,
	)
}

func NewMockConfiguration(path string) (*models.MockConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	config := &models.MockConfiguration{}
	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	if len(os.Args) < 4 {
		os.Exit(1)
	}

	mockConfig, err := NewMockConfiguration(os.Args[1])
	if err != nil {
		os.Exit(3)
	}

	gen, err := NewGenerator(os.Args[2], mockConfig)
	if err != nil {
		os.Exit(2)
	}
	outDir := os.Args[3]
	if err := gen.CreateTests(outDir); err != nil {
		os.Exit(3)
	}
}
