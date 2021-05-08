# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Binary {
<#
.SYNOPSIS
Create binary

.DESCRIPTION
Create/upload a new binary to Cumulocity

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/binaries_create

.EXAMPLE
PS> New-Binary -File $File

Upload a log file

.EXAMPLE
PS> New-Binary -File $File -Type "c8y_upload" -Data @{ c8y_Global = @{} }

Upload a config file and make it globally accessible for all users


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # File to be uploaded as a binary (required)
        [Parameter(Mandatory = $true)]
        [string]
        $File,

        # Custom type. If left blank, the MIME type will be detected from the file extension
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "binaries create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y binaries create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y binaries create $c8yargs
        }
    }

    End {}
}
