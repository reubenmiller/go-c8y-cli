# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentTenant {
<#
.SYNOPSIS
Get current tenant

.DESCRIPTION
Get the current tenant associated with the current session

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/currenttenant_get

.EXAMPLE
PS> Get-CurrentTenant

Get the current tenant (based on your current credentials)


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currenttenant get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.currentTenant+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currenttenant get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currenttenant get $c8yargs
        }
    }

    End {}
}
