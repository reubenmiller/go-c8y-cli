# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleReferenceCollectionFromGroup {
<#
.SYNOPSIS
Get collection of user role references from a group

.DESCRIPTION
Get collection of user role references from a group

.EXAMPLE
PS> Get-RoleReferenceCollectionFromGroup -Group $Group.id

Get a list of role references for a user group


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Group,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userRoles getRoleReferenceCollectionFromGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.roleReference+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Group `
            | Group-ClientRequests `
            | c8y userRoles getRoleReferenceCollectionFromGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Group `
            | Group-ClientRequests `
            | c8y userRoles getRoleReferenceCollectionFromGroup $c8yargs
        }
        
    }

    End {}
}
