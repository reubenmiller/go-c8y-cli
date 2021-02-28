# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SupportedMeasurements {
<#
.SYNOPSIS
Get supported measurements/s of a device

.DESCRIPTION
Returns a list of fragments (valueFragmentTypes) related to the device


.EXAMPLE
PS> Get-SupportedMeasurements -Device $device.id

Get the supported measurements of a device by name

.EXAMPLE
PS> Get-SupportedMeasurements -Device $device.id

Get the supported measurements of a device (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices getSupportedMeasurements"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices getSupportedMeasurements $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices getSupportedMeasurements $c8yargs
        }
        
    }

    End {}
}
