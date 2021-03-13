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
        [switch] $WithBinary,

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "")]
        [string]
        $ProcessingMode,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Don't prompt for confirmation
        [switch] $Force
    )

    Process {

        if ($null -ne $Device) {
            $iDevice = Expand-Device $Device
        } else {
            $iDevice = PSc8y\New-TestDevice -Force:$Force
        }
        
        # Fake device (if whatif prevented it from being created)
        if ($Dry -and $null -eq $iDevice) {
            $iDevice = @{ id = "12345" }
        }
        
        $c8yEvent = PSc8y\New-Event `
            -Device $iDevice.id `
            -Time:$Time `
            -Type "c8y_ci_TestEvent" `
            -Text "Test CI Event" `
            -ProcessingMode:$ProcessingMode `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force
        
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
