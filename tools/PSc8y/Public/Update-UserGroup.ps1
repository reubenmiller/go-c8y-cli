# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-UserGroup {
<#
.SYNOPSIS
Update user group

.DESCRIPTION
Update an existing user group

.LINK
c8y userGroups update

.EXAMPLE
PS> Update-UserGroup -Id $Group -Name "customGroup2"

Update a user group

.EXAMPLE
PS> Get-UserGroupByName -Name $Group.name | Update-UserGroup -Name "customGroup2"

Update a user group (using pipeline)


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

        # name
        [Parameter()]
        [string]
        $Name,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userGroups update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.group+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y userGroups update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y userGroups update $c8yargs
        }
        
    }

    End {}
}
