Function ContainInCollection {
    <#
    .SYNOPSIS
    Tests whether a collection contains the given ids

    .EXAMPLE
    $Responses | Should -ContainInCollection $device1, $device2

    #>
    [CmdletBinding()]
    Param(
        $ActualValue,
        $Expected,
        [switch]$Negate,
        $Because,
        $CallerSessionState
    )

    $ExpectedIDS = $Expected | ForEach-Object {
        if ("$_" -match "^\d+$") { $_ } else { $_.id }
    }
    
    $ActualIDS = if ("$ActualValue" -match "^\s*\{") {
        # Detected json input
        $ActualValue | ConvertFrom-Json -Depth 100 | ForEach-Object { $_.id }
    } else {
        $ActualValue | ForEach-Object {
            if ($_ -match "^\d+$") { $_ } else { $_.id }
        }
    }

    $compare = Compare-Object -ReferenceObject $ExpectedIDS -DifferenceObject $ActualIDS -ErrorAction SilentlyContinue
    [bool] $Pass = $null -eq $compare

    If ( $Negate ) { $Pass = -not($Pass) }

    if (-Not $Pass) {
        if ($Negate) {
            $FailureMessage = "Expected: collection not to contain ids: $($ExpectedIDS -join ',') but got $($ActualIDS -join ',')"
        } else {
            $FailureMessage = "Expected: collection to contain ids: $($ExpectedIDS -join ',') but got $($ActualIDS -join ',')"
        }
    }

    $ObjProperties = @{
        Succeeded      = $Pass
        FailureMessage = $FailureMessage
    }
    return New-Object PSObject -Property $ObjProperties
}
