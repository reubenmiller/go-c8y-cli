Function New-TestMeasurement {
<#
.SYNOPSIS
Create a new test measurement

.DESCRIPTION
Create a test measurement for a device.

If the device is not provided then a test device will be created automatically

.EXAMPLE
New-TestMeasurement

Create a new test device and then create a measurement on it

.EXAMPLE
New-TestMeasurement -Device "myExistingDevice"

Create a measurement on the existing device "myExistingDevice"
#>
    [cmdletbinding()]
    Param(
        # Device id, name or object. If left blank then a randomized device will be created
        [Parameter(
            Mandatory = $true,
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
        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Exclude Device -Command "measurements create"
        $Template = ""
        if (-Not $Template) {
            $Template = (Join-Path $script:Templates "test.measurement.jsonnet")
        }
        [void] $c8yargs.AddRange(@("--template", $Template))
    }

    Process {
        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y measurements create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y measurements create $c8yargs
        }
    }
}
