# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SystemOptionCollection {
<#
.SYNOPSIS
Get system option collection

.DESCRIPTION
Get a collection of system options. This endpoint provides a set of read-only properties pre-defined in platform configuration. The response format is exactly the same as for OptionCollection.

.LINK
c8y systemoptions list

.EXAMPLE
PS> Get-SystemOptionCollection

Get a list of system options


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "systemoptions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.optionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.option+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y systemoptions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y systemoptions list $c8yargs
        }
    }

    End {}
}
