Function New-TestEvent {
<#
.SYNOPSIS
Create a new test event

.DESCRIPTION
Create a test event for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestEvent

Create a new test device and then create an event on it

.EXAMPLE
New-TestEvent -Device "myExistingDevice"

Create an event on the existing device "myExistingDevice"
#>
    [cmdletbinding()]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [object] $Device,

        # Time
        [Parameter()]
        [string]
        $Time = "0s",

        # Add a dummy file to the event
        [switch] $WithBinary
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        $commonOptions = @{} + $PSBoundParameters
        $commonOptions.Remove("Device")
        $commonOptions.Remove("Time")
        $commonOptions.Remove("WithBinary")

        if ($null -ne $Device) {
            $iDevice = Expand-Device $Device
        } else {
            $iDevice = PSc8y\New-TestDevice @commonOptions
        }
        
        # Fake device (if Dry prevented it from being created)
        if ($Dry -and $null -eq $iDevice) {
            $iDevice = @{ id = "12345" }
        }
        
        $options = @{} + $PSBoundParameters
        $options["Device"] = $iDevice.id
        $options.Remove("WithBinary")
        $options["Type"] = "c8y_ci_TestEvent"
        $options["Text"] = "Test CI Event"
        
        $c8yEvent = PSc8y\New-Event @options
        
        if ($WithBinary) {
            if ($Dry -and $null -eq $iDevice) {
                $c8yEvent = @{ id = "12345" }
            }
            
            $tempfile = New-TemporaryFile
            "Cumulocity test content" | Out-File -LiteralPath $tempfile
            $null = PSc8y\New-EventBinary `
                -Id $c8yEvent.id `
                -File $tempfile `
                -Force:$Force
            
            Remove-Item $tempfile
        }
        
        $c8yEvent
    }
}
