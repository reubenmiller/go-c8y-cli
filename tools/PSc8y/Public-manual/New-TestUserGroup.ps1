Function New-TestUserGroup {
<#
.SYNOPSIS
Create a test user group

.DESCRIPTION
Create a new test user group using a random name

.EXAMPLE
New-TestUserGroup -Name mygroup

Create a new user group with the prefix "mygroup". A random postfix will be added to it
#>
    [cmdletbinding()]
    Param(
        # Name of the user group. A random postfix will be added to it to make it unique
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testgroup"
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        $options = @{} + $PSBoundParameters
        $options["Name"] = New-RandomString -Prefix "${Name}_"
        PSc8y\New-UserGroup @options
    }
}
