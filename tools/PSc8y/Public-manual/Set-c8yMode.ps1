Function Set-c8yMode {
    <#
    .SYNOPSIS
    Set cli mode temporarily

    .EXAMPLE
    Set-c8yMode -Mode dev

    Enable development mode (all command enabled) temporarily. The active session file will not be updated
    #>
    [cmdletbinding()]
    Param(
        # Mode
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [ValidateSet("dev", "qual", "prod")]
        [string] $Mode
    )

    c8y settings update --shell powershell mode $Mode | Out-String | Invoke-Expression
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Enabled $Mode mode (temporarily)" -ForegroundColor Green
    }
}
