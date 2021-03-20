# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-DeviceRequest {
<#
.SYNOPSIS
Delete device request

.DESCRIPTION
Delete an existing device registration request

.LINK
c8y devicecredentials deleteNewDeviceRequest

.EXAMPLE
PS> Remove-DeviceRequest -Id "$serial_91019192078"

Delete a new device request


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # New Device Request ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicecredentials deleteNewDeviceRequest"
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
            | c8y devicecredentials deleteNewDeviceRequest $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devicecredentials deleteNewDeviceRequest $c8yargs
        }
        
    }

    End {}
}
