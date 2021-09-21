# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-SoftwareVersion {
<#
.SYNOPSIS
Uninstall software version on a device

.DESCRIPTION
Uninstall software version on a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_uninstall

.EXAMPLE
PS> Remove-SoftwareVersion -Device $mo.id -Software go-c8y-cli -Version 1.0.0

Uninstall a software package version


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device or agent where the software should be installed
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Software name (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Software,

        # Software version
        [Parameter()]
        [object[]]
        $Version,

        # Software action
        [Parameter()]
        [ValidateSet('delete')]
        [string]
        $Action
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions uninstall"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y software versions uninstall $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y software versions uninstall $c8yargs
        }
        
    }

    End {}
}
