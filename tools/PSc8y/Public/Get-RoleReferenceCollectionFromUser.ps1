# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleReferenceCollectionFromUser {
<#
.SYNOPSIS
Get collection of user role references from a user

.DESCRIPTION
Get collection of user role references from a user

.EXAMPLE
PS> Get-RoleReferenceCollectionFromUser -User $User.id

Get a list of role references for a user


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $User,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userRoles getRoleReferenceCollectionFromUser"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.roleReference+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $User `
            | c8y userRoles getRoleReferenceCollectionFromUser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $User `
            | c8y userRoles getRoleReferenceCollectionFromUser $c8yargs
        }
        
    }

    End {}
}
