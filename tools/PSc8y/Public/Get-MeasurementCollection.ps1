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
        $Unit,

        # Maximum number of results
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateRange(1,2000)]
        [int]
        $PageSize,

        # Include total pages statistic
        [Parameter()]
        [switch]
        $WithTotalPages,

        # Get a specific page result
        [Parameter()]
        [int]
        $CurrentPage,

        # Maximum number of pages to retrieve when using -IncludeAll
        [Parameter()]
        [int]
        $TotalPages,

        # Include all results
        [Parameter()]
        [switch]
        $IncludeAll,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

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
        if ($PSBoundParameters.ContainsKey("PageSize")) {
            $Parameters["pageSize"] = $PageSize
        }
        if ($PSBoundParameters.ContainsKey("WithTotalPages") -and $WithTotalPages) {
            $Parameters["withTotalPages"] = $WithTotalPages
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        $Parameters["device"] = PSc8y\Expand-Id $Device

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "measurements" `
            -Verb "list" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.measurementCollection+json" `
            -ItemType "application/vnd.com.nsn.cumulocity.measurement+json" `
            -ResultProperty "measurements" `
            -Raw:$Raw `
            -CurrentPage:$CurrentPage `
            -TotalPages:$TotalPages `
            -IncludeAll:$IncludeAll
    }

    End {}
}
