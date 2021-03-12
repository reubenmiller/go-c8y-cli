# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Binary {
<#
.SYNOPSIS
Create binary

.DESCRIPTION
Create/upload a new binary to Cumulocity

.LINK
c8y binaries create

.EXAMPLE
PS> New-Binary -File $File

Upload a log file

.EXAMPLE
PS> New-Binary -File $File -Data @{ c8y_Global = @{}; type = "c8y_upload" }

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
        $File
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
