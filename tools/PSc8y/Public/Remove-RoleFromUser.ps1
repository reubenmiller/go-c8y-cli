# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-RoleFromUser {
<#
.SYNOPSIS
Unassign role from user

.DESCRIPTION
Unassign/delete role from a user

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/userroles_deleteRoleFromUser

.EXAMPLE
PS> Remove-RoleFromUser -User $User.id -Role "ROLE_MEASUREMENT_READ"

Remove a role from the given user


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $User,

        # Role name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Role,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userroles deleteRoleFromUser"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Role `
            | Group-ClientRequests `
            | c8y userroles deleteRoleFromUser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Role `
            | Group-ClientRequests `
            | c8y userroles deleteRoleFromUser $c8yargs
        }
        
    }

    End {}
}
