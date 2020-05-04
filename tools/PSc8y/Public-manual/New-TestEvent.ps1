Function New-TestEvent {
<#
.SYNOPSIS
Create a new test event
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [object] $Device,

        [switch] $WithBinary,

        [switch] $Force
    )

    if ($null -ne $Device) {
        $iDevice = Expand-Device $Device
    } else {
        $iDevice = PSc8y\New-TestDevice -Force:$Force
    }

    $Event = PSc8y\New-Event `
        -Device $iDevice.id `
        -Time "1970-01-01" `
        -Type "c8y_ci_TestEvent" `
        -Text "Test CI Event" `
        -Force:$Force

    if ($WithBinary) {
        $tempfile = New-TemporaryFile
        "Cumulocity test content" | Out-File -LiteralPath $tempfile
        $null = PSc8y\New-EventBinary `
            -Id $Event.id `
            -File $tempfile `
            -Force:$Force

        Remove-Item $tempfile
    }

    $Event
}
