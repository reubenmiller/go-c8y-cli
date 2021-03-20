# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-RoleToGroup {
<#
.SYNOPSIS
Add role to user group

.DESCRIPTION
Add a role to an existing user group

.LINK
c8y userroles addRoleToGroup

.EXAMPLE
PS> Add-RoleToGroup -Group "${NamePattern}*" -Role "*ALARM_*"

Add a role to a group using wildcards

.EXAMPLE
PS> Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*" | Add-RoleToGroup -Group "${NamePattern}*"

Add a role to a group using wildcards (using pipeline)


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

        # User role id (required)
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
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userroles addRoleToGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Role `
            | Group-ClientRequests `
            | c8y userroles addRoleToGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Role `
            | Group-ClientRequests `
            | c8y userroles addRoleToGroup $c8yargs
        }
        
    }

    End {}
}
