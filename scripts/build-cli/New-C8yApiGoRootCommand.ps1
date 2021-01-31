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

    $Name = $Specification.information.name

    if (!$Name) {
        Write-Error "Missing root command name"
        return
    }
    $NameCamel = $Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)
    $Description = $Specification.information.description
    $DescriptionLong = $Specification.information.descriptionLong

    $SubcommandsCode = New-Object System.Text.StringBuilder
    $RootImportCode = New-Object System.Text.StringBuilder

    $File = Join-Path -Path $OutputDir -ChildPath ("{0}RootCmd.go" -f $Name)

    foreach ($endpoint in $Specification.endpoints) {
        if ($endpoint.skip -eq $true) {
            Write-Verbose ("Skipping [{0}]" -f $endpoint.name)
            continue
        }
        $EndpointName = $endpoint.name
        $EndpointNameCamel = $EndpointName[0].ToString().ToUpperInvariant() + $EndpointName.Substring(1)

        $null = $SubcommandsCode.AppendLine("    cmd.AddCommand(New${EndpointNameCamel}Cmd().getCommand())")
    }

    # Create root import command helper
    $null = $RootImportCode.AppendLine("    // ${Name} commands")
    $null = $RootImportCode.AppendLine("    rootCmd.AddCommand(New${NameCamel}RootCmd().getCommand())")

    $Template = @"
package cmd

import (
    "github.com/spf13/cobra"
)

type ${NameCamel}Cmd struct {
    *baseCmd
}

func New${NameCamel}RootCmd() *${NameCamel}Cmd {
    ccmd := &${NameCamel}Cmd{}

    cmd := &cobra.Command{
        Use:   "${Name}",
        Short: "${Description}",
        Long:  ``${DescriptionLong}``,
    }

    // Subcommands
$SubcommandsCode

    ccmd.baseCmd = newBaseCmd(cmd)

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
