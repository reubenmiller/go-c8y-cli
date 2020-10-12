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
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testgroup",

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
        $GroupName = New-RandomString -Prefix "${Name}_"
        $TestGroup = PSc8y\New-Group `
            -Name $GroupName `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force
        $TestGroup
    }
}
