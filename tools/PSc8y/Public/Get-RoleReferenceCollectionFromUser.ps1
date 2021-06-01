# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleReferenceCollectionFromUser {
<#
.SYNOPSIS
Get role references from user

.DESCRIPTION
Get collection of user role references from a user

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/userroles_getRoleReferenceCollectionFromUser

.EXAMPLE
PS> Get-RoleReferenceCollectionFromUser -User $User.id

Get a list of role references for a user


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
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
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userroles getRoleReferenceCollectionFromUser"
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
            | Group-ClientRequests `
            | c8y userroles getRoleReferenceCollectionFromUser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $User `
            | Group-ClientRequests `
            | c8y userroles getRoleReferenceCollectionFromUser $c8yargs
        }
        
    }

    End {}
}
