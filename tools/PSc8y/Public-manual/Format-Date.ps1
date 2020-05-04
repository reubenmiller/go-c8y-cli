Function Format-Date {
<#
.SYNOPSIS
Gets a Cumulocity (ISO-8601) formatted DateTime string in the specified timezone

.DESCRIPTION
All Cumulocity REST API calls that require a date, must be in the ISO-8601 format. This function
allows the user to easily generate the correct format including the correct timezone information.

.NOTES
The standard powershell Get-Date does not have any timezone information.

.EXAMPLE
Format-Date

Get current datetime (now) as an ISO8601 formatted string

.EXAMPLE
[TimeZoneInfo]::GetSystemTimeZones() | Foreach-Object { Format-Date -Timezone $_ }

Get current datetime (now) as an ISO8601 formatted string in each of the timezones

.OUTPUTS
String
#>
    [CmdletBinding()]
    Param(
        # DateTime to be converted to ISO-8601 format. Accepts piped input
        [Parameter(Mandatory=$false,
                   Position = 0,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [datetime[]] $InputObject = @(),

        # Timezone to use when converting the DateTime object. Defaults to Local System Timezone
        [TimeZoneInfo] $TimeZone = $null
    )

    Begin {
        if ($null -eq $TimeZone) {
            $TimeZone = [TimeZoneInfo]::Local
        }

        $InputDates = New-Object System.Collections.ArrayList
    }

    Process {
        if ($null -ne $InputObject -and $InputObject.Count -ne 0) {
            $null = $InputDates.AddRange($InputObject)
        } else {
            $null = $InputDates.Add((Get-Date))
        }
    }

    End {
        foreach ($iDate in $InputDates) {

            # Get DateTime as ISO 8601 formatted string
            # It does not contain the timezone, this will have to be added manually
            $DateWithoutTimezone = Get-Date $iDate -Format "yyyy-MM-ddTHH:mm:ss.fff"

            # Get the time zone offset at a specific time
            $TimeZoneOffset = $TimeZone.GetUtcOffset($iDate)

            if (!($TimeZoneOffset.Hours -eq 0 -and $TimeZoneOffset.Minutes -eq 0)) {
                if ($TimeZoneOffset.TotalSeconds -ge 0) {
                    $OffsetStr = "+" + "$($TimeZoneOffset.Hours)".PadLeft(2, '0')
                } else {
                    # The minus sign is automatically in the hours.
                    $OffsetStr = "" + "$($TimeZoneOffset.Hours)".PadLeft(2, '0')
                }

                if ($TimeZoneOffset.Minutes -ne 0) {
                    # Math Absolute is required because when a timezone is before GMT, then the Minutes are negative
                    $OffsetStr += ":" + "$([math]::Abs($TimeZoneOffset.Minutes))".PadLeft(2, '0')
                }

                $DateWithTimezone = "${DateWithoutTimezone}${OffsetStr}"
            } else {
                # GMT 0 / UTC
                $DateWithTimezone = "${DateWithoutTimezone}Z"
            }

            $DateWithTimezone
        }
    }
}
