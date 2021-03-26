# Code generated from specification version 1.0.0: DO NOT EDIT
Function Approve-DeviceRequest {
<#
.SYNOPSIS
Approve device request

.DESCRIPTION
Approve a new device request. Note: a device can only be approved if the platform has received a request for device credentials.

.LINK
c8y deviceregistration approve

.EXAMPLE
PS> Approve-DeviceRequest -Id $DeviceRequest.id

Approve a new device request


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device identifier (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Status of registration
        [Parameter()]
        [ValidateSet('ACCEPTED')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceregistration approve"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.newDeviceRequest+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration approve $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration approve $c8yargs
        }
        
    }

    End {}
}
