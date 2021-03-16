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
    [cmdletbinding()]
    Param(
        # Name of the username. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testuser"
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "TemplateVars"
    }

    Process {
        $Username = New-RandomString -Prefix "${Name}_"
        $options = @{} + $PSBoundParameters
        $options.Remove("Name")
        $options["UserName"] = $Username
        $options["Password"] = New-RandomString
        
        PSc8y\New-User @options
    }
}
