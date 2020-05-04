Function New-TestOperation {
    <#
.SYNOPSIS
Create a new test operation
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [object] $Device,

        [switch] $Force
    )

    if ($null -ne $Device) {
        $iAgent = Expand-Device $Device
    }
    else {
        $iAgent = PSc8y\New-TestAgent -Force:$Force
    }

    PSc8y\New-Operation `
        -Device $iAgent.id `
        -Description "Test operation" `
        -Data @{
        c8y_Restart = @{
                parameters = @{ }
            }
        } `
        -Force:$Force
}
