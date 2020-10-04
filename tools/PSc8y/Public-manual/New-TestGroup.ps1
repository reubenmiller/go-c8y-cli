Function New-TestGroup {
<#
.SYNOPSIS
Create a test user group

.DESCRIPTION
Create a new test user group using a random name

.EXAMPLE
New-TestGroup -Name mygroup

Create a new user group with the prefix "mygroup". A random postfix will be added to it
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Name of the user group. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testgroup",

        # Don't prompt for confirmation
        [switch] $Force
    )

    $GroupName = New-RandomString -Prefix "${Name}_"
    $TestGroup = PSc8y\New-Group -Name $GroupName -Force:$Force

    $TestGroup
}
