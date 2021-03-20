# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserGroupByName {
<#
.SYNOPSIS
Get user group by name

.DESCRIPTION
Get an existing user group by name

.LINK
c8y usergroups getByName

.EXAMPLE
PS> Get-UserGroupByName -Name $Group.name

Get user group by its name


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
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "usergroups getByName"
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
            | c8y usergroups getByName $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y usergroups getByName $c8yargs
        }
        
    }

    End {}
}
