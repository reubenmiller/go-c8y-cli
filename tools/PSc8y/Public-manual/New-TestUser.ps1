Function New-TestUser {
<#
.SYNOPSIS
Create a new test user

.DESCRIPTION
Create a user with a randomized username

.EXAMPLE
New-TestUser

Create a new test user

.EXAMPLE
New-TestUser -Name "myExistingDevice"

Create a new test user with a custom username prefix
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Name of the username. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testuser",

        # Don't prompt for confirmation
        [switch] $Force
    )

    $Username = New-RandomString -Prefix "${Name}_"

    PSc8y\New-User -UserName $Username -Password (New-RandomString) -Force:$Force
}
