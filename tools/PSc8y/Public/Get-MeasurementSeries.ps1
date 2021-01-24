# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MeasurementSeries {
<#
.SYNOPSIS
Get a collection of measurements based on filter parameters

.DESCRIPTION
Get a collection of measurements based on filter parameters

.EXAMPLE
PS> Get-MeasurementSeries -Device $Device.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"

Get a list of measurements for a particular device

.EXAMPLE
PS> Get-MeasurementSeries -Device $Measurement2.source.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"

Get measurement series c8y_Temperature.T on a device

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-MeasurementSeries -Series "c8y_Temperature.T"

Get measurement series from a device (using pipeline)


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

        # measurement type and series name, e.g. c8y_AccelerationMeasurement.acceleration
        [Parameter()]
        [string[]]
        $Series,

        # Fragment name from measurement.
        [Parameter()]
        [ValidateSet('DAILY','HOURLY','MINUTELY')]
        [string]
        $AggregationType,

        # Start date or date and time of measurement occurrence. Defaults to last 7 days
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of measurement occurrence. Defaults to the current time
        [Parameter()]
        [string]
        $DateTo,

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
        if ($PSBoundParameters.ContainsKey("Series")) {
            $Parameters["series"] = $Series
        }
        if ($PSBoundParameters.ContainsKey("AggregationType")) {
            $Parameters["aggregationType"] = $AggregationType
        }
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
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
            -Verb "getSeries" `
            -Parameters $Parameters `
            -Type "application/json" `
            -ItemType "" `
            -ResultProperty "" `
            -Raw:$Raw `
            -CurrentPage:$CurrentPage `
            -TotalPages:$TotalPages `
            -IncludeAll:$IncludeAll
    }

    End {}
}
