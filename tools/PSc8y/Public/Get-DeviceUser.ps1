# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceUser {
<#
.SYNOPSIS
Get device user

.DESCRIPTION
Retrieve the device owner's username and state (enabled or disabled) of a specific managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_user_get

.EXAMPLE
PS> Get-DeviceUser -Id $device.id

Get device user by id

.EXAMPLE
PS> Get-DeviceUser -Id $device.name

Get device user by name


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices user get"
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
            | c8y devices user get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices user get $c8yargs
        }
        
    }

    End {}
}
