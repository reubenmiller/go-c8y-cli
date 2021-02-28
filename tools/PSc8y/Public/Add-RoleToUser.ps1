# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-RoleToUser {
<#
.SYNOPSIS
Add role to a user

.DESCRIPTION
Add role to a user

.LINK
c8y userRoles addRoleTouser

.EXAMPLE
PS> Add-RoleToUser -User $User.id -Role "ROLE_ALARM_READ"

Add a role (ROLE_ALARM_READ) to a user

.EXAMPLE
PS> Add-RoleToUser -User "customUser_*" -Role "*ALARM_*"

Add a role to a user using wildcards

.EXAMPLE
PS> Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*" | Add-RoleToUser -User "customUser_*"

Add a role to a user using wildcards (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User prefix or full username (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $User,

        # User role id
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Role,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userRoles addRoleTouser"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Role)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Role `
            | Group-ClientRequests `
            | c8y userRoles addRoleTouser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Role `
            | Group-ClientRequests `
            | c8y userRoles addRoleTouser $c8yargs
        }
        
    }

    End {}
}
