# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DeviceService {
<#
.SYNOPSIS
Create service

.DESCRIPTION
Create a new service which is attached to the given device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_services_create

.EXAMPLE
PS> New-DeviceService -Id $software.id -Data "custom.value=test" -Global -ChildType addition

Create a new service for a device (as a child addition)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

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
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices services create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices services create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices services create $c8yargs
        }
        
    }

    End {}
}
