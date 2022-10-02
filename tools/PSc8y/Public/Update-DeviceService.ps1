# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceService {
<#
.SYNOPSIS
Update service status

.DESCRIPTION
Update service status

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_services_update

.EXAMPLE
PS> Update-DeviceService -Id 12345 -Status up

Update service status


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Service id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Service name
        [Parameter()]
        [string]
        $Name,

        # Service type, e.g. systemd
        [Parameter()]
        [string]
        $ServiceType,

        # Service status
        [Parameter()]
        [ValidateSet('up','down','unknown')]
        [string]
        $Status
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices services update"
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
            | c8y devices services update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices services update $c8yargs
        }
        
    }

    End {}
}
