# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObjectCollection {
<#
.SYNOPSIS
Get managed object collection

.DESCRIPTION
Get a collection of managedObjects based on filter parameters

.LINK
c8y inventory list

.EXAMPLE
PS> Get-ManagedObjectCollection

Get a list of managed objects

.EXAMPLE
PS> Get-ManagedObjectCollection -Ids $Device1.id, $Device2.id

Get a list of managed objects by id


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # List of ids.
        [Parameter()]
        [string[]]
        $Ids,

        # ManagedObject type.
        [Parameter()]
        [string]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

        # managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z).
        [Parameter()]
        [string]
        $Text,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents,

        # Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved
        [Parameter()]
        [switch]
        $SkipChildrenNames
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y inventory list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y inventory list $c8yargs
        }
    }

    End {}
}
