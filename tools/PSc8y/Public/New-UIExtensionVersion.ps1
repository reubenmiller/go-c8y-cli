# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-UIExtensionVersion {
<#
.SYNOPSIS
Create a new version of an extension

.DESCRIPTION
Uploaded version and tags can only contain upper and lower case letters, integers and ., +, -. Other characters are prohibited.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_extensions_versions_create

.EXAMPLE
PS> New-UIExtensionVersion -Extension 1234 -File ./myapp.zip -Version "2.0.0"

Create a new version for an extension


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Extension
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Extension,

        # The ZIP file to be uploaded
        [Parameter()]
        [string]
        $File,

        # The JSON file with version information. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Version,

        # The JSON file with version information. todo (required)
        [Parameter(Mandatory = $true)]
        [string[]]
        $Tags
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui extensions versions create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Extension `
            | Group-ClientRequests `
            | c8y ui extensions versions create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Extension `
            | Group-ClientRequests `
            | c8y ui extensions versions create $c8yargs
        }
        
    }

    End {}
}
