# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceUser {
<#
.SYNOPSIS
Update device user

.DESCRIPTION
Update the device owner's state (enabled or disabled) of a specific managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_user_update

.EXAMPLE
PS> Update-DeviceUser -Id $device.id -Enabled

Enable a device user

.EXAMPLE
PS> Update-DeviceUser -Id $device.name -Enabled:$false

Disable a device user


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

        # Specifies if the device's owner is enabled or not.
        [Parameter()]
        [switch]
        $Enabled
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices user update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedobjectuser+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices user update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices user update $c8yargs
        }
        
    }

    End {}
}
