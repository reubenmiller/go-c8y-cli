# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-BinaryCollection {
<#
.SYNOPSIS
Get binary collection

.DESCRIPTION
Get a collection of inventory binaries. The results include the meta information about binary and not the binary itself.


.LINK
c8y binaries list

.EXAMPLE
PS> Get-BinaryCollection -PageSize 100

Get a list of binaries


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "binaries list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y binaries list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y binaries list $c8yargs
        }
    }

    End {}
}
