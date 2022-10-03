# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceServiceCollection {
<#
.SYNOPSIS
Get device services collection

.DESCRIPTION
Get a collection of services of a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_services_list

.EXAMPLE
PS> Get-DeviceServiceCollection -Device 12345

Get services for a specific device

.EXAMPLE
PS> Get-Device -Id 12345 | Get-DeviceServiceCollection

Get services for a specific device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device id (required for name lookup) (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Additional query filter
        [Parameter()]
        [string]
        $Query,

        # Filter by service type e.g. systemd
        [Parameter()]
        [string]
        $ServiceType,

        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by service status
        [Parameter()]
        [ValidateSet('up','down','unknown')]
        [string]
        $Status,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices services list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices services list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices services list $c8yargs
        }
        
    }

    End {}
}
