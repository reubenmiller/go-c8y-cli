Function New-RandomString {
    [cmdletbinding()]
    Param(
        [string] $Prefix,

        [string] $Postfix
    )
    $RandomPart = -join ((48..57) + (97..122) |
        Get-Random -Count 10 |
        ForEach-Object { [char]$_ })

    Write-Output "${Prefix}${RandomPart}${Postfix}"
}
