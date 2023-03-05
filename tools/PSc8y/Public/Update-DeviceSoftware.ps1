# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceSoftware {
<#
.SYNOPSIS
Update service status

.DESCRIPTION
Update service status

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_software_update

.EXAMPLE
PS> Update-DeviceSoftware -Id 12345 -Status up

Update software status


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Service name
        [Parameter()]
        [string]
        $Name,

        # Software version
        [Parameter()]
        [string]
        $Version,

        # Software url
        [Parameter()]
        [string]
        $Url,

        # Software type, e.g. apt
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices software update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices software update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices software update $c8yargs
        }
        
    }

    End {}
}
