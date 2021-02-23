# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceRequestCollection {
<#
.SYNOPSIS
Get a collection of new device requests

.DESCRIPTION
Get a collection of device requests

.EXAMPLE
PS> Get-DeviceRequestCollection

Get a list of new device requests


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceCredentials listNewDeviceRequests"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.newDeviceRequestCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.newDeviceRequest+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y deviceCredentials listNewDeviceRequests $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y deviceCredentials listNewDeviceRequests $c8yargs
        }
    }

    End {}
}
