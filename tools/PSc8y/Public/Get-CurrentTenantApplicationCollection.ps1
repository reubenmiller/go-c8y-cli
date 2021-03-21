# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentTenantApplicationCollection {
<#
.SYNOPSIS
List applications in current tenant

.DESCRIPTION
Get the applications of the current tenant

.LINK
c8y currenttenant listApplications

.EXAMPLE
PS> Get-CurrentTenantApplicationCollection

Get a list of applications in the current tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currenttenant listApplications"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.currentTenant+json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currenttenant listApplications $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currenttenant listApplications $c8yargs
        }
    }

    End {}
}
