# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FirmwareCollection {
<#
.SYNOPSIS
Get firmware collection

.DESCRIPTION
Get a collection of firmware packages (managedObjects) based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_list

.EXAMPLE
PS> Get-FirmwareCollection

Get a list of firmware packages


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y firmware list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y firmware list $c8yargs
        }
    }

    End {}
}
