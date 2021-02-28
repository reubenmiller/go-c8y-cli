# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-User {
<#
.SYNOPSIS
Update user

.DESCRIPTION
Update properties, reset password or enable/disable for a user in a tenant

.EXAMPLE
PS> Update-User -Id $User.id -FirstName "Simon"

Update a user


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

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

        # User activation status (true/false)
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
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.user+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Id)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y users update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y users update $c8yargs
        }
        
    }

    End {}
}
