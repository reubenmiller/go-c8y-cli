# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MeasurementSeries {
<#
.SYNOPSIS
Get measurement series

.DESCRIPTION
Get a collection of measurements based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/measurements_getSeries

.EXAMPLE
PS> Get-MeasurementSeries -Device $Device.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01"

Get a list of measurements for a particular device

.EXAMPLE
PS> Get-MeasurementSeries -Device $Measurement2.source.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01"

Get measurement series c8y_Temperature.T on a device

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-MeasurementSeries -Series "c8y_Temperature.T"

Get measurement series from a device (using pipeline)

.EXAMPLE
PS> Get-MeasurementSeries -Device 12345 -Series app_Weather.barometer -AggregationFunction avg -AggregationInterval 1w

Get a list of series and calculate the average and use 1 week intervals

.EXAMPLE
PS> Get-MeasurementSeries -Device 12345 -Series app_Weather.temperature -AggregationFunction avg,count,sum -AggregationInterval 1h

Get a list of series and calculate the average, count and sum and use 1 hour intervals


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

        # measurement type and series name, e.g. c8y_AccelerationMeasurement.acceleration
        [Parameter()]
        [string[]]
        $Series,

        # (time series only) Selects aggregation functions that are calculated for each selected aggregation interval
        [Parameter()]
        [ValidateSet('min','max','avg','sum','count','stdDevPop','stdDevSamp')]
        [string[]]
        $AggregationFunction,

        # (time series only) Fetch results are aggregated using a time interval specified by an integer followed by a unit. Available units of (s)econd, (m)inute, (h)our, (d)ay, week, (M)onth, (q)uarter and (y)ear
        [Parameter()]
        [string[]]
        $AggregationInterval,

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
        $DateTo
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "measurements getSeries"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y measurements getSeries $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y measurements getSeries $c8yargs
        }
        
    }

    End {}
}
