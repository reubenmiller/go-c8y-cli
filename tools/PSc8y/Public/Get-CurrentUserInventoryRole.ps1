# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentUserInventoryRole {
<#
.SYNOPSIS
Get a specific inventory role of the current user

.DESCRIPTION
Get a specific inventory role of the current user

.EXAMPLE
PS> Get-CurrentUserInventoryRoleCollection -PageSize 1 | Get-CurrentUserInventoryRole

Get an inventory role of the current user (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Role id. Note: lookup by name is not yet supported (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users getCurrentUserInventoryRole"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventoryrole+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y users getCurrentUserInventoryRole $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y users getCurrentUserInventoryRole $c8yargs
        }
        
    }

    End {}
}
