Function New-TestUser {
<#
.SYNOPSIS
Create a new test user
#>
    [cmdletbinding()]
    Param(
        # Name of the username. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testuser",

        [switch] $Force
    )

    $Username = New-RandomString -Prefix "${Name}_"

    PSc8y\New-User -UserName $Username -Password (New-RandomString) -Force:$Force
}
