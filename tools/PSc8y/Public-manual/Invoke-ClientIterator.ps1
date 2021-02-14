function Invoke-ClientIterator {
<#
.SYNOPSIS
Convert the input objects into a format that can be easily piped to the c8y binary directly

.DESCRIPTION
Calling the go c8y directly involves converting powershell objects either into json lines, or
just passing on the id.

The iterator can also also format the input data and fan it out (turning 1 input item into x items) by using the -Format and -Repeat parameters respectively.


.EXAMPLE
Get-DeviceCollection | Invoke-ClientIterator | c8y devices update --data "mytype=myNewTypeValue"
Get a collection of devices and add a fragment "mytype: 'myNewTypeValue'" to each device.

.EXAMPLE
Get-Device myDeviceName | Invoke-ClientIterator -Repeat 5 | c8y measurements create --template example.jsonnet

Lookup a device by its name and then create 5 measurements using a jsonnet template

.EXAMPLE
@(1..20) | Invoke-ClientIterator "device" | c8y devices create --template example.jsonnet

Create 20 devices naming from "device0001" to "device0020" using a jsonnet template file.

.EXAMPLE
@(1..2) | Invoke-ClientIterator "device_{0}-{1}" -Repeat 2 | c8y devices create

Create 4 (Input count x Repeat) devices with the following names.

```powershell
device_1-0
device_1-1
device_2-0
device_2-1
```

.EXAMPLE
@(1..2) | Invoke-ClientIterator "device_{0}-{2}" -Repeat 2 | c8y devices create

Create 4 (Input count x Repeat) devices with the following names (using 1-indexed values when repeating)

```powershell
device_1-1
device_1-2
device_2-1
device_2-2
```

#>
    [CmdletBinding(
        DefaultParameterSetName = "string"
    )]
    param (
        # Input objects to be piped to native c8y binary
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            ValueFromRemainingArguments = $true,
            Mandatory = $true,
            Position = 1
        )]
        [object[]]
        $InputObject,

        # Format string to be applied to each value. The format string is $Format -f $item
        # The value will be prefixed to the input objects by default. However the format string
        # can be customized by using a powershell string format (i.e. "{0:00}" )
        #
        # Other format variables (additional )
        # "{0}" is the current input object (i.e. {0:000} for 0 padded numbers)
        # "{1}" is the repeat counter from 0..Repeat-1
        # "{2}" is the repeat counter from 1..Repeat
        [Parameter(
            Position = 0,
            ParameterSetName = "string"
        )]
        [string]
        $Format,

        # Repeat each input x times. Useful when wanting to use the same item in multiple commands.
        # If a value less than 1 is provided, then it will be set to 1 automatically
        [Parameter(
            Position = 2
        )]
        [int]
        $Repeat,

        # Convert the items to json lines
        [Parameter(
            ParameterSetName = "json"
        )]
        [switch]
        $AsJSON
    )

    begin {
        if ($Repeat -lt 1) {
            $Repeat = 1
        }
        $ValueFormatter = "${Format}{0:0000}"
        if ($Format.Contains("{") -and $Format.Contains("}")) {
            $ValueFormatter = $Format
        }
    }

    process {
        if ($PSCmdlet.ParameterSetName -eq "json") {
            foreach ($item in ($InputObject)) {
                
                $OutputItem = if ($item -is [string] -or $item -is [int]) {
                    @{id=$item}
                } else {
                    $item
                }

                for ($i = 0; $i -lt $Repeat; $i++) {
                    Write-Output (ConvertTo-Json $OutputItem -Depth 100 -Compress)
                }
            }
        } else {
            foreach ($item in (Expand-Id $InputObject)) {
                for ($i = 0; $i -lt $Repeat; $i++) {
                    Write-Output ($ValueFormatter -f $item, $i, ($i+1))
                }
            }
        }
        
    }
}
