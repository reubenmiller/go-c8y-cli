Function Format-ConfirmationMessage {
<#
.SYNOPSIS
Format the confirmation message from a cmdlet name and input object
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0)]
        [string] $Name,

        [Parameter(
            Mandatory = $true,
            Position = 1)]
        [AllowNull()]
        [object] $inputObject,

        [string] $IgnorePrefix = ""
    )

    $parts = New-Object System.Collections.ArrayList;

    # Remove fully qualified module name
    $Name = $Name -replace "^\w+\\", ""

    foreach ($item in ($Name -csplit '(?=[A-Z\-])')) {
        if ($item -eq "-" -or $item -eq "" -or $item -eq $IgnorePrefix) {
            continue;
        }
        if ($parts.Count -eq 0) {
            $null = $parts.Add($item);
        } else {
            $null = $parts.Add("$item".ToLowerInvariant());
        }
    }

    if ($inputObject.id -and $inputObject.name) {
        $null = $parts.Add(("[{1} ({0})]" -f $inputObject.id, $inputObject.name))
    } elseif ($inputObject.id) {
        $null = $parts.Add(("[{0}]" -f $inputObject.id))
    } elseif ($inputObject) {
        $null = $parts.Add("[{0}]" -f $inputObject)
    } else {
        # Don't add anything
    }

    $parts -join " "
}
