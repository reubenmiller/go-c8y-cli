Function Get-ClientSetting {
    <#
    .SYNOPSIS
    Get the Cumulocity binary settings
    
    .DESCRIPTION
    Get the Cumulocity binary settings which used by the cli tool
    
    .EXAMPLE
    Get-ClientSetting
    
    Show the current c8y cli tool settings
    #>
    [cmdletbinding()]
    Param()

    $c8ybinary = Get-ClientBinary
    $settings = & $c8ybinary settings list --pretty=false
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Client error when getting settings"
        return
    }

    $JSONArgs = @{}
    if ($PSVersionTable.PSVersion.Major -gt 5) {
        $JSONArgs.Depth = 100
    }
    $settings | ConvertFrom-JSON @JSONArgs
}
