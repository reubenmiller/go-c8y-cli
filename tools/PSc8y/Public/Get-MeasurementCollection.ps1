# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MeasurementCollection {
<#
.SYNOPSIS
Get measurement collection

.DESCRIPTION
Get a collection of measurements based on filter parameters

.LINK
c8y measurements list

.EXAMPLE
PS> Get-MeasurementCollection

Get a list of measurements

.EXAMPLE
PS> Get-MeasurementCollection -Device $Device.id -Type "TempReading"

Get a list of measurements for a particular device

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-MeasurementCollection

Get measurements from a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Measurement type.
        [Parameter()]
        [string]
        $Type,

        # value fragment type
        [Parameter()]
        [string]
        $ValueFragmentType,

        # value fragment series
        [Parameter()]
        [string]
        $ValueFragmentSeries,

        # Fragment name from measurement (deprecated).
        [Parameter()]
        [string]
        $FragmentType,

        # Start date or date and time of measurement occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of measurement occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # Return the newest instead of the oldest measurements. Must be used with dateFrom and dateTo parameters
        [Parameter()]
        [switch]
        $Revert,

        # Results will be displayed in csv format. Note: -IncludeAll, is not supported when using using this parameter
        [Parameter()]
        [switch]
        $CsvFormat,

        # Results will be displayed in Excel format Note: -IncludeAll, is not supported when using using this parameter
        [Parameter()]
        [switch]
        $ExcelFormat,

        # Every measurement fragment which contains 'unit' property will be transformed to use required system of units.
        [Parameter()]
        [ValidateSet('imperial','metric')]
        [string]
        $Unit
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "measurements list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.measurementCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.measurement+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y measurements list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y measurements list $c8yargs
        }
        
    }

    End {}
}
