Function Open-Website {
<# 
.SYNOPSIS
Open a browser to the cumulocity website

.DESCRIPTION
Opens the default web browser to the Cumulocity application or directly to a device page in the Device Management application

.NOTES
When running on Linux, it relies on xdg-open. If it is not found, then only the URL will be printed to the console.
The user can then try to open the URL by clicking on the link if they are using a modern terminal which supports url links.

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
    # Application to open
    [ValidateSet("cockpit", "administration", "devicemanagement", "fieldbus4")]
    [Parameter(
      ParameterSetName="Application",
      Position=0
    )]
    [string] $Application = "cockpit",

    # Name of the device to open in devicemanagement. Only the first matching device will be used to open the c8y website.
    [Parameter(
      ParameterSetName="Device",
      Position=0,
      ValueFromPipelineByPropertyName=$true,
      ValueFromPipeline=$true
    )]
    [object[]] $Device,

    # Device page to open
    [Parameter(
      ParameterSetName="Device",
      Position=1
    )]
    [ValidateSet("device-info", "measurements", "alarms", "control", "events", "service_monitoring", "identity")]
    [string] $Page = "control",

    # Browser to use to open the webpage
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
        if ($null -eq $Device) {
          $Url = "/apps/devicemanagement/index.html"
        } else {
          $DeviceInfo = Expand-Device $Device | Select-Object -First 1
          
          if (!$DeviceInfo) {
            Write-Error "Could not find a matching devices to [$Device]"
            return;
          }
          $Url = "/apps/devicemanagement/index.html#/device/{0}/{1}" -f @($DeviceInfo.id, $Page)
        }
        break;
      }
    }

    $Url = (Get-C8ySessionProperty -Name "host") + $Url

    if ($Url -notmatch "https?://") {
      $Url = "https://$Url"
    }

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
        if ($IsMacOS) {
          $null = Start-Process "open" $Url -PassThru
        } elseif ($IsLinux) {
          if (Get-Command "xdg-open" -ErrorAction SilentlyContinue) {
            $null = Start-Process "xdg-open" $Url -PassThru
          } else {
            Write-Warning "xdg-open is not present on your system. Try clicking on the URL to open it in a browser (if supported by your console)"
          }
        } else {
          $null = Start-Process $Browser $Url -PassThru
        }
      }
    }
  }
}
