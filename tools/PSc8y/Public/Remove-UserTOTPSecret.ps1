# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-UserTOTPSecret {
<#
.SYNOPSIS
Revoke a user's TOTP (TFA) secret

.DESCRIPTION
Revoke/delete a user's TOTP (TFA) secret to force them to setup TFA again.

This is required when the user loses their TFA configuration, or it is compromised.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/users_revokeTOTPSecret

.EXAMPLE
PS> Remove-UserTFA -Id $User.id

Revoke a user's TOTP (TFA) secret


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users revokeTOTPSecret"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y users revokeTOTPSecret $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y users revokeTOTPSecret $c8yargs
        }
        
    }

    End {}
}
