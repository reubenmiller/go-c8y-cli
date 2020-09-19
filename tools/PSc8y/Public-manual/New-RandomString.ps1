Function New-RandomString {
<# 
.SYNOPSIS
Create a random string

.DESCRIPTION
Helper utility to quickly create a randomized string which can be used
when adding unique names to devices or another other properties

Note: It should not be used for encryption!

.EXAMPLE
New-RandomString -Prefix "hello_"

Create a random string with the "hello" prefix. i.e `hello_jta6fzwvo7`

.EXAMPLE
New-RandomString -Postfix "_device"

Create a random string which ends with "_device", i.e. `1qs7mc2o3t_device`

#>
    [cmdletbinding()]
    Param(
        # Prefix to be added before the random string
        [string] $Prefix,

        # Postfix to be added after the random string
        [string] $Postfix
    )
    $RandomPart = -join ((48..57) + (97..122) |
        Get-Random -Count 10 |
        ForEach-Object { [char]$_ })

    Write-Output "${Prefix}${RandomPart}${Postfix}"
}
