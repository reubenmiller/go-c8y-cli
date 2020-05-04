Function Open-Website {
<# 
.SYNOPSIS
Open a browser to the cumulocity website

.PARAMETER Application
Application to open

.PARAMETER Device
Name of the device to open in devicemanagement. Only the first matching device will be used to open the c8y website.

.PARAMETER Page
Device page to open

.PARAMETER Browser
Browser to use to open the webpage

.EXAMPLE
Open-Website -Application "cockpit"

Open the cockpit application

.EXAMPLE
Open-Website -Device myDevice01

Open the devicemanagement to the device (default) control page for myDevice01

.EXAMPLE
Open-Website -Device myDevice01 -Page alarms

Open the devicemanagement to the device alarm page for myDevice01

#>
  [cmdletbinding(
    DefaultParameterSetName="Device"
  )]
  Param(
    [ValidateSet("cockpit", "administration", "devicemanagement", "fieldbus4")]
    [Parameter(
      ParameterSetName="Application",
      Position=0
    )]
    [string] $Application = "cockpit",

    [Parameter(
      ParameterSetName="Device",
      Position=0,
      ValueFromPipelineByPropertyName=$true,
      ValueFromPipeline=$true
    )]
    [object[]] $Device,

    [Parameter(
      ParameterSetName="Device",
      Position=1
    )]
    [ValidateSet("device-info", "measurements", "alarms", "control", "events", "service_monitoring", "identity")]
    [string] $Page = "control",

    [ValidateSet("chrome", "firefox", "ie", "edge")]
    [string] $Browser = "chrome"
  )
  Process {
    switch ($PSCmdlet.ParameterSetName) {
      "Application" {
        $Url = "/apps/{0}/index.html" -f $Application
        break;
      }

      "Device" {
        $DeviceInfo = Expand-Device $Device | Select-Object -First 1

        if (!$DeviceInfo) {
          Write-Error "Could not find a matching devices to [$Device]"
          return;
        }
        $Url = "/apps/devicemanagement/index.html#/device/{0}/{1}" -f @($DeviceInfo.id, $Page)
        break;
      }
    }

    # todo: add expand uri function to c8y binary
    $Url = (Get-C8ySessionProperty -Name "host") + $Url

    # Print a link to the console, so the user can click on it
    Write-Host "Open page: $Url" -ForegroundColor Gray

    switch ($Browser) {
      "ie" {
        # Open the url in the default browser...for the people who use Internet Explorer
        # it is most likely still their default browser so it should work.
        $null = Start-Process $Url -PassThru
      }
      "edge" {
        $null = Start-Process "microsoft-edge:$Url" -PassThru -ErrorAction SilentlyContinue
      }
      Default {
        $null = Start-Process $Browser $Url -PassThru
      }
    }
  }
}
