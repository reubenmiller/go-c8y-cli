Function ContainRequest {
    <#
    .SYNOPSIS
    Tests whether a value contains a specific rest requets or not

    .EXAMPLE
    $Responses | Should -ContainRequest "WHATIF GET /event/events" -Total 1

    .EXAMPLE
    $Responses | Should -ContainRequest "WHATIF GET /event/events" -Total 0

    No matching requests to /event/events should be present

    #>
    [CmdletBinding()]
    Param(
        $ActualValue,
        $Request,
        $Total = 1,
        $Minimum = 0,
        $Maximum = 0,
        [switch]$Negate,
        $Because,
        $CallerSessionState
    )
    [array] $Requests = $ActualValue | ForEach-Object {
        if ($_ -match "Sending request:\s*(\w+)\s+(.+)") {
            $method = $Matches[1]
            $url = $Matches[2] -replace "https?://[^\/]+", ""
            "$method $url"
            
        } elseif ($_ -match "What If: Sending\s*\[(\w+)\] request to \[(.+?)\]") {
            $method = $Matches[1]
            $url = $Matches[2] -replace "https?://[^\/]+", ""
            "WHATIF $method $url"
        }
    }

    $UsingTotal = $PSBoundParameters.ContainsKey("Total")
    $RangeMessage = "$Minimum-$Maximum"
    if ($UsingTotal) {
        $Minimum = $Total
        $Maximum = $Total
        $RangeMessage = "$Total"
    }
    [bool] $Pass = $false

    if ($Requests.Count -eq 0) {
        if ($Total -eq 0) {
            $Pass = $true
        } else {
            $FailureMessage = 'Expected: requests to include some requests but 0 where found'
        }
    } else {
        $pattern = [regex]::Escape($Request)
        $TotalMatches = @($Requests -match $pattern)
        [bool]$Pass = $TotalMatches.Count -le $Maximum -and $TotalMatches.Count -ge $Minimum
        If ( $Negate ) { $Pass = -not($Pass) }
        
        If ( -not($Pass) ) {
            If ( $Negate ) {
                $FailureMessage = 'Expected: requests `n{{{0}}} to not contain {{{1}}} but matches were found.' -f @(
                    ($Prefix + ($Requests -join "$Prefix") + $Suffix),
                    $Request
                )
            }
            Else {
                $Prefix = ""
                $Suffix = ""
                if ($Requests.Count -gt 1) {
                    $Prefix = "`n  "
                    $Suffix = "`n"
                }
                $FailureMessage = 'Expected: requests {{{0}}} to contain {{{1}}} request/s, but found {{{2}}} matching {{{3}}}' -f @(
                    ($Prefix + ($Requests -join "$Prefix") + $Suffix),
                    $RangeMessage,
                    $TotalMatches.Count,
                    $Request
                )
            }
        }
    }

    $ObjProperties = @{
        Succeeded      = $Pass
        FailureMessage = $FailureMessage
    }
    return New-Object PSObject -Property $ObjProperties
}
