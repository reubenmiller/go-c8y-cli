# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-UserGroup {
<#
.SYNOPSIS
Delete user group

.DESCRIPTION
Delete an existing user group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/usergroups_delete

.EXAMPLE
PS> Remove-UserGroup -Id $Group.id

Delete a user group

.EXAMPLE
PS> Get-UserGroupByName -Name $Group.name | Remove-UserGroup

Delete a user group (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "usergroups delete"
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
            | c8y usergroups delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y usergroups delete $c8yargs
        }
        
    }

    End {}
}
