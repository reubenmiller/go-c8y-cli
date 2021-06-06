# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-UserFromGroup {
<#
.SYNOPSIS
Delete user from group

.DESCRIPTION
Delete an existing user from a user group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/userreferences_deleteUserFromGroup

.EXAMPLE
PS> Remove-UserFromGroup -Group $Group.id -User $User.id

Delete a user from a user group


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group ID (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # User id/username (required)
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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userreferences deleteUserFromGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $User `
            | Group-ClientRequests `
            | c8y userreferences deleteUserFromGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $User `
            | Group-ClientRequests `
            | c8y userreferences deleteUserFromGroup $c8yargs
        }
        
    }

    End {}
}
