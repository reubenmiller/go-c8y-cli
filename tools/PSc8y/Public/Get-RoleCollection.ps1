# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleCollection {
<#
.SYNOPSIS
Get collection of user roles

.DESCRIPTION
Get collection of user roles

.EXAMPLE
PS> Get-RoleCollection -PageSize 100

Get a list of roles


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userRoles list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.role+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            ,(c8y userRoles list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions)
        }
        else {
            c8y userRoles list $c8yargs
        }
    }

    End {}
}
