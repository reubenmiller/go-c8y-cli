# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-DeviceSoftware {
<#
.SYNOPSIS
Delete service

.DESCRIPTION
Delete an existing service

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_software_delete

.EXAMPLE
PS> Remove-DeviceSoftware -Id 12345 -Name ntp

Remove software


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

        # Software name
        [Parameter()]
        [string]
        $Name,

        # Software version
        [Parameter()]
        [string]
        $Version
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices software delete"
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
            | c8y devices software delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices software delete $c8yargs
        }
        
    }

    End {}
}
