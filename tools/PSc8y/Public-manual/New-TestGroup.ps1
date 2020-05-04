Function New-TestGroup {
<#
.SYNOPSIS
Create a test user group
#>
    [cmdletbinding()]
    Param(
        # Name of the user group. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testgroup",

        [switch] $Force
    )

    $GroupName = New-RandomString -Prefix "${Name}_"
    $TestGroup = PSc8y\New-Group -Name $GroupName -Force:$Force

    $TestGroup
}
