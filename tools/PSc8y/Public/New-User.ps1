# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-User {
<#
.SYNOPSIS
Create user

.DESCRIPTION
Create a new user so that they can access the tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/users_create

.EXAMPLE
PS> New-user -Username "$Username" -Email "testuser@no-reply.dummy.com" -Password "$NewPassword"

Create a user


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User name, unique for a given domain. Max: 1000 characters
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $UserName,

        # User first name
        [Parameter()]
        [string]
        $FirstName,

        # User last name
        [Parameter()]
        [string]
        $LastName,

        # User phone number. Format: '+[country code][number]', has to be a valid MSISDN
        [Parameter()]
        [string]
        $Phone,

        # User email address
        [Parameter()]
        [string]
        $Email,

        # User activation status (true/false)
        [Parameter()]
        [switch]
        $Enabled,

        # User password. Min: 6, max: 32 characters. Only Latin1 chars allowed
        [Parameter()]
        [string]
        $Password,

        # Send password reset email to the user instead of setting a password
        [Parameter()]
        [ValidateSet('true','false')]
        [switch]
        $SendPasswordResetEmail,

        # Custom properties to be added to the user
        [Parameter()]
        [object]
        $CustomProperties,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.user+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $UserName `
            | Group-ClientRequests `
            | c8y users create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $UserName `
            | Group-ClientRequests `
            | c8y users create $c8yargs
        }
        
    }

    End {}
}
