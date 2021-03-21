Function New-TestAlarm {
<#
.SYNOPSIS
Create a new test alarm

.DESCRIPTION
Create a test alarm for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestAlarm

Create a new test device and then create an alarm on it

.EXAMPLE
New-TestAlarm -Device "myExistingDevice"

Create an alarm on the existing device "myExistingDevice"
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
        [object] $Device
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "TemplateVars"
    }
    Begin {
        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "alarms create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDevice+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
        [void] $c8yargs.AddRange(@(
            "--template",
            "test.alarm.jsonnet"
        ))
    }

    Process {
        if ($null -ne $Device) {
            $iDevice = Expand-Device $Device
        } else {
            $iDevice = PSc8y\New-TestDevice -Force:$Force -AsPSObject
        }

        # Fake device (if Dry prevented it from being created)
        if ($Dry -and $null -eq $iDevice) {
            $iDevice = @{ id = "12345" }
        }

        if ($ClientOptions.ConvertToPS) {
            $iDevice.id `
            | c8y alarms create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $iDevice.id `
            | c8y alarms create $c8yargs
        }
    }
}
