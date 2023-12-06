function Group-ClientRequests {
<# 
.SYNOPSIS
Groups the input into array of a given maximum size.

.DESCRIPTION
Groups the input into array of a given maximum size. It will pass the piped input as array rather than individual items
This cmdlet is mainly used internally by the module.

.EXAMPLE
PS> $Id | Group-ClientRequests | c8y devices delete $c8yargs

Group and normalize the input objects to be compatible with piping to the native c8y binary
#>
    [CmdletBinding()]
    param (
        # Input objects to be piped to native c8y binary
        [AllowNull()]
        [AllowEmptyCollection()]
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            ValueFromRemainingArguments = $true,
            Mandatory = $true,
            Position = 0
        )]
        [object[]]
        $InputObject,

        # Grouping size
        [int]
        $Size = 2000,

        # Output objects as PSObjects rather than json text
        [switch]
        $AsPSObject
    )

    begin {
        $Buffer = New-Object System.Collections.ArrayList
    }

    process {
        foreach ($item in $InputObject) {
            if ($AsPSObject -or $item -is [string] -or $item -is [int] -or $item -is [long]) {
                $pipeitem = $item
            } else {
                $pipeitem = ConvertTo-Json -InputObject $item -Depth 100 -Compress
            }
            [void]$Buffer.Add($pipeitem)

            if ($Buffer.Count -eq $Size) {
                $b = $Buffer;
                $Buffer = New-Object System.Collections.ArrayList($Size)
                ,$b
            }
        }
    }

    end {
        if ($Buffer.Count -ne 0) {  
            ,$Buffer
        }
    }
}
