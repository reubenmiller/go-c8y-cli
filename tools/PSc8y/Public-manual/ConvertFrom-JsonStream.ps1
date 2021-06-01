Function ConvertFrom-JsonStream {
    <# 
.SYNOPSIS
Convert json text to powershell objects as the objects are piped to it

.DESCRIPTION
The cmdlet will convert each input as a separate json line, and it will convert it as soon as it is
received in the pipeline (instead of waiting for the entire input). 

Each input should contain a single json object.

.NOTES
ConvertFrom-Json can not be used for streamed json as it waits to receive all piped input before it starts trying
to convert the json to PowerShell objects.

.EXAMPLE
Get-DeviceCollection | Get-Device -AsPSObject:$false | ConvertFrom-JsonStream

Convert the pipeline objects to json as they come through the pipeline.

#>
    [CmdletBinding()]
    param(
        # Input json lines
        [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline = $true)]
        [AllowNull()]
        [object[]]
        ${InputObject},
    
        # Maximum object depth to allow
        [ValidateRange(1, 2147483647)]
        [int]
        ${Depth} = 100,

        # Convert json to a hashtable instead of a PSCustom Object
        [switch]
        $AsHashtable
    )    
    process {
        foreach ($item in $inputObject) {
            # Strip color codes (if present)
            $item = $item -replace '\x1b\[[0-9;]*m'
            Write-Output (ConvertFrom-Json $item -Depth:$Depth -AsHashtable:$AsHashtable) -NoEnumerate
        }
    }
}
