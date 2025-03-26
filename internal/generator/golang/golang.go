package golang

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
)

//go:embed command.tmpl
var TemplateCommand string

type TemplateContext struct {
	Context Command
}

type Command struct {
	InitBody                 bool
	Use                      string
	PackageName              string
	Name                     string
	RESTMethod               string
	RESTPath                 string
	GetBodyContents          string
	PrepareRequest           string
	RESTBodyBuilderOptions   string
	PostActionOptions        string
	Examples                 string
	PreRunFunction           string
	FlagBuilderOptions       string
	RestHeaderBuilderOptions []string
	Spec                     *models.Command
}

func NewContextFromSpecification(commandSpec *models.Command) *Command {
	cmd := Command{
		Use:                      commandSpec.Alias.Go,
		Name:                     commandSpec.Name,
		RESTMethod:               commandSpec.Method,
		RESTPath:                 commandSpec.Path,
		Spec:                     commandSpec,
		Examples:                 commandSpec.GetExamples(),
		RestHeaderBuilderOptions: make([]string, 0),
	}

	// Pre run function
	switch commandSpec.GetMethod() {
	case "POST":
		cmd.PreRunFunction = "f.CreateModeEnabled()"
	case "PUT":
		cmd.PreRunFunction = "f.UpdateModeEnabled()"
	case "DELETE":
		cmd.PreRunFunction = "f.DeleteModeEnabled()"
	default:
		cmd.PreRunFunction = "nil"
	}

	if (strings.EqualFold(cmd.RESTMethod, "PUT") || strings.EqualFold(cmd.RESTMethod, "POST")) && len(commandSpec.Body) > 0 {
		cmd.InitBody = true
	}

	// Add common parameters
	switch commandSpec.GetMethod() {
	case "POST", "PUT", "DELETE":
		cmd.FlagBuilderOptions += "flags.WithProcessingMode(),"
	}

	//
	// Headers
	//
	if commandSpec.ContentType != "" {
		cmd.RestHeaderBuilderOptions = append(cmd.RestHeaderBuilderOptions, fmt.Sprintf(`flags.WithStaticStringValue("Content-Type", "%s"),`, commandSpec.ContentType))
	}
	if commandSpec.Accept != "" {
		cmd.RestHeaderBuilderOptions = append(cmd.RestHeaderBuilderOptions, fmt.Sprintf(`flags.WithStaticStringValue("Accept", "%s"),`, commandSpec.Accept))
	}

	switch commandSpec.GetMethod() {
	case "POST", "PUT", "DELETE":
		cmd.RestHeaderBuilderOptions = append(cmd.RestHeaderBuilderOptions, "flags.WithProcessingModeValue(),")
	}

	return &cmd
}

func GetAllParameters(c *models.Command) []models.Parameter {
	return c.GetAllParameters()
}

func Generate(w io.Writer, ctx TemplateContext) error {
	funcs := sprig.FuncMap()
	funcs["GetAllParameters"] = GetAllParameters
	funcs["toCobraFlag"] = GetFlagCode
	funcs["toCobraGetter"] = GetFlagGetterCode

	// commandSpec.GetAllParameters()
	t, err := template.New("command").Funcs(funcs).Parse(TemplateCommand)
	if err != nil {
		return err
	}
	err = t.Execute(w, ctx)
	return err
}
