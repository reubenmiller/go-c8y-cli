# Code generated from specification version 1.0.0: DO NOT EDIT
Function Set-DeviceRequiredAvailability {
<#
.SYNOPSIS
Set required availability

.DESCRIPTION
Set the required availability of a device. Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_availability_set

.EXAMPLE
PS> Set-DeviceRequiredAvailability -Id $device.id -Interval 10

Set the required availability of a device by name to 10 minutes

.EXAMPLE
PS> Get-ManagedObject -Id $device.id | PSc8y\Set-DeviceRequiredAvailability -Interval 10

Set the required availability of a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Interval in minutes (required)
        [Parameter(Mandatory = $true)]
        [long]
        $Interval
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices availability set"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices availability set $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices availability set $c8yargs
        }
        
    }

    End {}
}
