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

    $Name = $Specification.group.name.ToLower()
    $BaseName = Split-Path -Path $Specification.group.name.ToLower() -Leaf

    if ($Specification.group.skip -eq $true) {
        Write-Information "Specification is marked to be ignored"
        return
    }

    if (!$Name) {
        Write-Error "Missing root command name"
        return
    }
    $BaseNameLowercase = $BaseName.ToLower()
    $NameCamel = $BaseNameLowercase[0].ToString().ToUpperInvariant() + $BaseNameLowercase.Substring(1)
    $Description = $Specification.group.description
    $DescriptionLong = $Specification.group.descriptionLong
    $Hidden = $Specification.group.hidden

    $SubcommandsCode = New-Object System.Text.StringBuilder
    $RootImportCode = New-Object System.Text.StringBuilder
    $GoImports = New-Object System.Text.StringBuilder
    $CommandOptions =  New-Object System.Text.StringBuilder

    $File = Join-Path -Path $OutputDir -ChildPath ("{0}.auto.go" -f $BaseNameLowercase)

    foreach ($endpoint in $Specification.commands) {
        if ($endpoint.skip -eq $true) {
            Write-Verbose ("Skipping [{0}]" -f $endpoint.name)
            continue
        }
        $EndpointName = $endpoint.name
        $GoCmdName = $endpoint.alias.go
        $GoCmdNameLower = $GoCmdName.ToLower() -replace "-", "_"
        $GoCmdNameCamel = ($GoCmdName[0].ToString().ToUpperInvariant() + $GoCmdName.Substring(1)) -replace '-(\p{L})', { $_.Groups[1].Value.ToUpper() }
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
