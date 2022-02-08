# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceAvailability {
<#
.SYNOPSIS
Get device availability

.DESCRIPTION
Retrieve the date when a specific managed object (by a given ID) sent the last message to Cumulocity IoT.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_availability_get

.EXAMPLE
PS> Get-DeviceAvailability -Id $Device.id

Get a device's availability by id

.EXAMPLE
PS> Get-DeviceAvailability -Id $Device.name

Get a device's availability by name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices availability get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices availability get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices availability get $c8yargs
        }
        
    }

    End {}
}
