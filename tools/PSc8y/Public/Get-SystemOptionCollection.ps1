# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SystemOptionCollection {
<#
.SYNOPSIS
Get collection of system options

.DESCRIPTION
This endpoint provides a set of read-only properties pre-defined in platform configuration. The response format is exactly the same as for OptionCollection.

.EXAMPLE
PS> Get-SystemOptionCollection

Get a list of system options


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
        $Parameters = @{}

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "systemOptions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.optionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.option+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y systemOptions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y systemOptions list $c8yargs
        }
    }

    End {}
}
