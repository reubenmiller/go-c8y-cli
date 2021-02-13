function ConvertTo-ClientPipeline {
<#
.SYNOPSIS
Convert the input objects into a format that can be easily piped to the c8y binary directly

.DESCRIPTION
Calling the go c8y directly involves converting powershell objects either into json lines, or
just passing on the id.

.EXAMPLE
Get-DeviceCollection | ConvertTo-ClientPipeline | c8y devices update --data "mytype=myNewTypeValue"
Get a collection of devices and add a fragment "mytype: 'myNewTypeValue'" to each device.

.EXAMPLE
Get-Device myDeviceName | ConvertTo-ClientPipeline -Repeat 5 | c8y measurements create --template example.jsonnet

Lookup a device by its name and then create 5 measurements using a jsonnet template
#>
    [CmdletBinding()]
    param (
        # Input objects to be piped to native c8y binary
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromRemainingArguments = $true,
            Mandatory = $true
        )]
        [object[]]
        $InputObject,

        # Repeat each input x times. Useful when wanting to use the same item in multiple commands
        [int]
        $Repeat
    )

    begin {
        if ($Repeat -lt 1) {
            $Repeat = 1
        }
    }

    process {
        foreach ($item in (Expand-Id $InputObject)) {
            for ($i = 0; $i -lt $Repeat; $i++) {
                Write-Output $item
            }
        }
    }
}
