# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SmartGroupCollection {
<#
.SYNOPSIS
List smart group collection

.DESCRIPTION
Get a collection of smart groups based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/smartgroups_list

.EXAMPLE
PS> Get-SmartGroupCollection

Get a list of smart groups

.EXAMPLE
PS> Get-SmartGroupCollection -Name "$Name*"

Get a list of smart groups with the names starting with 'myText'


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y smartgroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y smartgroups list $c8yargs
        }
    }

    End {}
}
