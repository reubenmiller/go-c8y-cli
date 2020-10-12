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
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testuser",

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Don't prompt for confirmation
        [switch] $Force
    )

    Process {
        $Username = New-RandomString -Prefix "${Name}_"
        
        PSc8y\New-User `
            -UserName $Username `
            -Password (New-RandomString) `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force
    }
}
