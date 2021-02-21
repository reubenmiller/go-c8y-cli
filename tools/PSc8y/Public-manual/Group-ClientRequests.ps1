function Group-ClientRequests {
    [CmdletBinding(
    )]
    param (
        # Input objects to be piped to native c8y binary
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            ValueFromRemainingArguments = $true,
            Mandatory = $true,
            Position = 0
        )]
        [object[]]
        $InputObject,

        [switch]
        $AsPSObject
    )

    begin {
        $items = New-Object System.Collections.ArrayList
    }

    process {
        foreach ($item in $InputObject) {
            if ($AsPSObject -or $item -is [string]) {
                $pipeitem = $item 
            } else {
                $pipeitem = ConvertTo-Json -InputObject $item -Depth 100 -Compress
            }
            $null = $items.Add($pipeitem)
        }
    }

    end {
        Write-Output $items -NoEnumerate
    }
}
