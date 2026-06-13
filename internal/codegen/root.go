package codegen

import (
	"strings"
)

// GenerateRootCommand ports New-C8yApiGoRootCommand.ps1: it renders the
// unformatted source of the <group>.auto.go root command file that registers
// all subcommands of a group.
func GenerateRootCommand(spec *Value) (fileName string, src []byte) {
	group := spec.Get("group")
	name := goPackageName(group.Get("name").Str())
	base := baseName(strings.ToLower(group.Get("name").Str()))
	useName := strings.ToLower(baseName(group.Get("name").Str()))

	baseNameLowercase := strings.ReplaceAll(base, "-", "_")
	nameCamel := upperFirst(baseNameLowercase)
	description := group.Get("description").Str()
	descriptionLong := group.Get("descriptionLong").Str()

	fileName = baseNameLowercase + ".auto.go"

	subcommandsCode := &strings.Builder{}
	goImports := &strings.Builder{}

	for _, endpoint := range spec.Get("commands").Items() {
		if strings.EqualFold(endpoint.Get("skip").Str(), "true") {
			continue
		}
		goCmdName := endpoint.Get("alias").Get("go").Str()
		goCmdNameLower := goPackageName(goCmdName)
		goCmdNameCamel := goCamelCase(goCmdName)
		importAlias := "cmd" + goCmdNameCamel

		goImports.WriteString(importAlias + " \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/" + name + "/" + goCmdNameLower + "\"\n")
		subcommandsCode.WriteString("    cmd.AddCommand(" + importAlias + ".New" + goCmdNameCamel + "Cmd(f).GetCommand())\n")
	}

	commandOptions := ""
	if group.Get("hidden").Truthy() {
		commandOptions = "\t\tHidden: true,\n"
	}

	b := &strings.Builder{}
	b.WriteString("package " + baseNameLowercase + "\n")
	b.WriteString("\n")
	b.WriteString("import (\n")
	b.WriteString("    \"github.com/spf13/cobra\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil\"\n")
	b.WriteString("    " + goImports.String() + "\n")
	b.WriteString(")\n")
	b.WriteString("\n")
	b.WriteString("type SubCmd" + nameCamel + " struct {\n")
	b.WriteString("    *subcommand.SubCommand\n")
	b.WriteString("}\n")
	b.WriteString("\n")
	b.WriteString("func NewSubCommand(f *cmdutil.Factory) *SubCmd" + nameCamel + " {\n")
	b.WriteString("    ccmd := &SubCmd" + nameCamel + "{}\n")
	b.WriteString("\n")
	b.WriteString("    cmd := &cobra.Command{\n")
	b.WriteString("        Use:   \"" + useName + "\",\n")
	b.WriteString("        Short: \"" + description + "\",\n")
	// The PowerShell template appends "`n" + $CommandOptions unconditionally
	// (an empty StringBuilder is still truthy in the if expression).
	b.WriteString("        Long:  `" + descriptionLong + "`,\n" + commandOptions + "\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    // Subcommands\n")
	b.WriteString(subcommandsCode.String() + "\n")
	b.WriteString("\n")
	b.WriteString("    ccmd.SubCommand = subcommand.NewSubCommand(cmd)\n")
	b.WriteString("\n")
	b.WriteString("    return ccmd\n")
	b.WriteString("}\n")
	b.WriteString("\n")

	return fileName, []byte(b.String())
}
