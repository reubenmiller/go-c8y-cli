# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-UserGroup {
<#
.SYNOPSIS
Create user group

.DESCRIPTION
Create a new user group

.LINK
c8y userGroups create

.EXAMPLE
PS> New-UserGroup -Name "$GroupName"

Create a user group


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userGroups create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.group+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y userGroups create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y userGroups create $c8yargs
        }
        
    }

    End {}
}
