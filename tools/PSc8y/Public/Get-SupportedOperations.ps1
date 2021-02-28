# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SupportedOperations {
<#
.SYNOPSIS
Get supported operations of a device

.DESCRIPTION
Returns a list of supported operations (fragments) for a device. The supported fragments list is returned from the c8y_SupportedOperations fragment of the device managed object.


.EXAMPLE
PS> Get-SupportedOperations -Device $device.id

Get the supported operations of a device by name

.EXAMPLE
PS> Get-SupportedOperations -Device $device.id

Get the supported operations of a device (using pipeline)


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices getSupportedOperations"
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
            | c8y devices getSupportedOperations $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices getSupportedOperations $c8yargs
        }
        
    }

    End {}
}
