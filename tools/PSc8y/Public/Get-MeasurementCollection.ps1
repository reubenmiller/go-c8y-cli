# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MeasurementCollection {
<#
.SYNOPSIS
Get a collection of measurements based on filter parameters

.DESCRIPTION
Get a collection of measurements based on filter parameters

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
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("ValueFragmentType")) {
            $Parameters["valueFragmentType"] = $ValueFragmentType
        }
        if ($PSBoundParameters.ContainsKey("ValueFragmentSeries")) {
            $Parameters["valueFragmentSeries"] = $ValueFragmentSeries
        }
        if ($PSBoundParameters.ContainsKey("FragmentType")) {
            $Parameters["fragmentType"] = $FragmentType
        }
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
        }
        if ($PSBoundParameters.ContainsKey("Revert")) {
            $Parameters["revert"] = $Revert
        }
        if ($PSBoundParameters.ContainsKey("CsvFormat")) {
            $Parameters["csvFormat"] = $CsvFormat
        }
        if ($PSBoundParameters.ContainsKey("ExcelFormat")) {
            $Parameters["excelFormat"] = $ExcelFormat
        }
        if ($PSBoundParameters.ContainsKey("Unit")) {
            $Parameters["unit"] = $Unit
        }

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
            | c8y measurements list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | c8y measurements list $c8yargs
        }
        
    }

    End {}
}
