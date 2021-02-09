Function ContainRequest {
    <#
    .SYNOPSIS
    Tests whether a value contains a specific rest requets or not
    #>
    [CmdletBinding()]
    Param(
        $ActualValue,
        $Request,
        $Total = 1,
        [switch]$Negate,
        $CallerSessionState
    )
    [array] $Requests = $ActualValue | Where-Object { $_ -match "Sending.*request to" }
    [bool] $Pass = $False

    if ($Requests.Count -eq 0) {
        $FailureMessage = 'Expected: value {{{0}}} does not contain any requests' -f ($ActualValue)
    } else {
        [array] $RequestParts = "$Request".Split(" ", 2)
        $RequestType = ".*"
        $RequestUrl = ""
        switch ($RequestParts.Count) {
            1 {
                $RequestUrl = $RequestParts[0]
                break
            }
            2 {
                $RequestType = $RequestParts[0]
                $RequestUrl = $RequestParts[1]
                break
            }
        }
        $pattern = "\[$RequestType\].*" + [regex]::Escape($RequestUrl) + ".*"

        [bool]$Pass = ($Requests -match $pattern).Count -eq $Total
        If ( $Negate ) { $Pass = -not($Pass) }
        
        If ( -not($Pass) ) {
            If ( $Negate ) {
                $FailureMessage = 'Expected: value {{{0}}} to not contain {{{1}}} but it was found.' -f ($Requests -join ","), $Request
            }
            Else {
                $FailureMessage = 'Expected: value {{{0}}} to contain {{{1}}} request/s matching {{{2}}}' -f ($Requests -join ","), $Total, $Request
            }
        }
    }

    $ObjProperties = @{
        Succeeded      = $Pass
        FailureMessage = $FailureMessage
    }
    return New-Object PSObject -Property $ObjProperties
}
