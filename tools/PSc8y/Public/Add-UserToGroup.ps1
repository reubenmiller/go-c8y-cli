# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-UserToGroup {
<#
.SYNOPSIS
Add user to group

.DESCRIPTION
Add an existing user to a group

.LINK
c8y userReferences addUserToGroup

.EXAMPLE
PS> Add-UserToGroup -Group $Group.id -User $User.id

Add a user to a user group


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

        # User id (required)
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
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userReferences addUserToGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.userReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $User `
            | Group-ClientRequests `
            | c8y userReferences addUserToGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $User `
            | Group-ClientRequests `
            | c8y userReferences addUserToGroup $c8yargs
        }
        
    }

    End {}
}
