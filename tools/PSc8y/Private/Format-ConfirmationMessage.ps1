Function Format-ConfirmationMessage {
<#
.SYNOPSIS
Format the confirmation message from a cmdlet name and input object
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            Position = 0)]
        [string] $Name,

        [Parameter(
            Mandatory = $true,
            Position = 1)]
        [AllowNull()]
        [object[]] $inputObject,

        [string] $IgnorePrefix = ""
    )

    Process {
    
        $parts = New-Object System.Collections.ArrayList;

        # Remove fully qualified module name
        $Name = $Name -replace "^\w+\\", ""

        foreach ($item in ($Name -csplit '(?=[A-Z\-])')) {
            if ($item -eq "-" -or $item -eq "" -or $item -eq $IgnorePrefix) {
                continue;
            }
            $item = $item -replace "^-", ""
            if ($parts.Count -eq 0) {
                $null = $parts.Add($item);
            } else {
                $null = $parts.Add("$item".ToLowerInvariant());
            }
        }

        foreach ($item in $inputObject) {
            if ($item -is [string]) {
                if ($item.StartsWith("{")) {
                    $item = ConvertFrom-Json $item -Depth 100 -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
                }
            }
            if ($item.id -and $item.name) {
                $null = $parts.Add(("[{1} ({0})]" -f $item.id, $item.name))
            } elseif ($item.id) {
                $null = $parts.Add(("[{0}]" -f $item.id))
            } elseif ($item) {
                $null = $parts.Add("[{0}]" -f $item)
            } else {
                # Don't add anything
            }
        }

        $parts -join " "
    }
}
