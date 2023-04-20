# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Measurement {
<#
.SYNOPSIS
Create measurement

.DESCRIPTION
Create a new measurement

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/measurements_create

.EXAMPLE
PS> New-Measurement -Device {{ randomdevice }} -Time "0s" -Type "myType" -Data @{ c8y_Winding = @{ temperature = @{ value = 25.0; unit = "°C" } } }

Create measurement


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The ManagedObject which is the source of this measurement.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Time of the measurement. Defaults to current timestamp.
        [Parameter()]
        [string]
        $Time,

        # The most specific type of this entire measurement.
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "measurements create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.measurement+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y measurements create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y measurements create $c8yargs
        }
        
    }

    End {}
}
