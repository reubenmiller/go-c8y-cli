Function New-C8yApiGoRootCommand {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [object] $Specification,

        [string] $OutputDir = "./"
    )

    $Name = $Specification.information.name.ToLower()
    $BaseName = Split-Path -Path $Specification.information.name.ToLower() -Leaf

    if (!$Name) {
        Write-Error "Missing root command name"
        return
    }
    $BaseNameLowercase = $BaseName.ToLower()
    $NameCamel = $BaseNameLowercase[0].ToString().ToUpperInvariant() + $BaseNameLowercase.Substring(1)
    $Description = $Specification.information.description
    $DescriptionLong = $Specification.information.descriptionLong
    $Hidden = $Specification.information.hidden

    $SubcommandsCode = New-Object System.Text.StringBuilder
    $RootImportCode = New-Object System.Text.StringBuilder
    $GoImports = New-Object System.Text.StringBuilder
    $CommandOptions =  New-Object System.Text.StringBuilder

    $File = Join-Path -Path $OutputDir -ChildPath ("{0}.auto.go" -f $BaseNameLowercase)

    foreach ($endpoint in $Specification.endpoints) {
        if ($endpoint.skip -eq $true) {
            Write-Verbose ("Skipping [{0}]" -f $endpoint.name)
            continue
        }
        $EndpointName = $endpoint.name
        $GoCmdName = $endpoint.alias.go
        $GoCmdNameLower = $GoCmdName.ToLower()
        $GoCmdNameCamel = $GoCmdName[0].ToString().ToUpperInvariant() + $GoCmdName.Substring(1)
        $ImportAlias = "cmd" + $GoCmdNameCamel

        $null = $GoImports.AppendLine("$ImportAlias `"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/$Name/$GoCmdNameLower`"")
        $null = $SubcommandsCode.AppendLine("    cmd.AddCommand(${ImportAlias}.New${GoCmdNameCamel}Cmd(f).GetCommand())")
    }

    # Create root import command helper
    $null = $RootImportCode.AppendLine("    // ${Name} commands")
    $null = $RootImportCode.AppendLine("    rootCmd.AddCommand(New${NameCamel}RootCmd(f).GetCommand())")

    
    if ($Hidden) {
        $null = $CommandOptions.AppendLine("		Hidden: true,")
    }

    $Template = @"
package $BaseNameLowercase

import (
    "github.com/spf13/cobra"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
    $GoImports
)

type SubCmd${NameCamel} struct {
    *subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmd${NameCamel} {
    ccmd := &SubCmd${NameCamel}{}

    cmd := &cobra.Command{
        Use:   "${BaseNameLowercase}",
        Short: "${Description}",
        Long:  ``${DescriptionLong}``,$(
            if ($CommandOptions) {
                "`n" + $CommandOptions
            }
        )
    }

    // Subcommands
$SubcommandsCode

    ccmd.SubCommand = subcommand.NewSubCommand(cmd)

    return ccmd
}

"@

    # Must not include BOM!
	$Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
	[System.IO.File]::WriteAllLines($File, $Template, $Utf8NoBomEncoding)

	# Auto format code
    $fmtErrors = & gofmt -w $File

    if ($LASTEXITCODE -ne 0) {
        Write-Error "gofmt errors. $fmtErrors"
    }

    # Return the import statements
    $RootImportCode.ToString()
}
