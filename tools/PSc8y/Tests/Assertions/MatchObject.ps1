Function MatchObject {
    <#
    .SYNOPSIS
    Tests whether the outputs matches the given object (root level key/values only)

    .EXAMPLE
    $Responses | Should -MatchObject @{name="12345"}

    #>
    [CmdletBinding()]
    Param(
        $ActualValue,
        $Expected,
        [switch]$Negate,
        [string[]] $Property,
        $Because,
        $CallerSessionState
    )

    $ActualValueCopy = $ActualValue
    if ($Property.Count -gt 0) {
        $ActualValueCopy | Select-Object -Property $Property
    }

    if ($ActualValueCopy -is [hashtable]) {
        $ActualKeys = $ActualValueCopy.Keys | Sort-Object
    } else {
        $ActualKeys = $ActualValueCopy.psobject.Properties.Name | Sort-Object
    }
    $ActualValues = $ActualKeys | ForEach-Object { $ActualValueCopy."$_" }

    if ($Expected -is [hashtable]) {
        $ExpectedKeys = $Expected.Keys | Sort-Object
    } else {
        $ExpectedKeys = $Expected.psobject.Properties.Name | Sort-Object
    }
    $ExpectedValues = $ExpectedKeys | ForEach-Object { $Expected."$_" }
    
    # compare keys
    $Compare = Compare-Object $ExpectedKeys $ActualKeys
    if ($null -ne $Compare) {
        [PSCustomObject]@{
            Succeeded = $false
            FailureMessage = "Object does not contain expected keys. Actual {{{0}}}, got {{{1}}}" -f @(
                ($ActualKeys -join ","),
                ($ExpectedKeys -join ",")
            )
        }
        return
    }

    # compare keys
    $Compare = Compare-Object $ExpectedKeys $ActualKeys
    if ($null -ne $Compare) {
        [PSCustomObject]@{
            Succeeded = $false
            FailureMessage = "Object does not contain expected values. Actual {{{0}}}, got {{{1}}}" -f @(
                ($ActualValues -join ","),
                ($ExpectedValues -join ",")
            )
        }
        return
    }

    [PSCustomObject]@{
        Succeeded = $true
        FailureMessage = ""
    }
    return    
}
