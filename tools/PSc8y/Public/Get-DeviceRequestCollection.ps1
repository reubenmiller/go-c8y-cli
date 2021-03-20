# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceRequestCollection {
<#
.SYNOPSIS
Get device request collection

.DESCRIPTION
Get a collection of device registration requests

.LINK
c8y devicecredentials listNewDeviceRequests

.EXAMPLE
PS> Get-DeviceRequestCollection

Get a list of new device requests


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicecredentials listNewDeviceRequests"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.newDeviceRequestCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.newDeviceRequest+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicecredentials listNewDeviceRequests $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicecredentials listNewDeviceRequests $c8yargs
        }
    }

    End {}
}
