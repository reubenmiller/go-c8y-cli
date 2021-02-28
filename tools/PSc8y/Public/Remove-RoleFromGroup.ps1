# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-RoleFromGroup {
<#
.SYNOPSIS
Unassign/Remove role from a group

.DESCRIPTION
Unassign/Remove role from a group

.EXAMPLE
PS> Remove-RoleFromGroup -Group $UserGroup.id -Role "ROLE_MEASUREMENT_READ"

Remove a role from the given user group


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # Role name, e.g. ROLE_TENANT_MANAGEMENT_ADMIN (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userRoles deleteRoleFromGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
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
            | c8y userRoles deleteRoleFromGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Role `
            | Group-ClientRequests `
            | c8y userRoles deleteRoleFromGroup $c8yargs
        }
        
    }

    End {}
}
