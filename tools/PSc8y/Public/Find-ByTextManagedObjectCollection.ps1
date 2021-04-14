# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-ByTextManagedObjectCollection {
<#
.SYNOPSIS
Find managed object by text collection

.DESCRIPTION
Find a collection of managedObjects which match a given text value

.LINK
c8y inventory findByText

.EXAMPLE
PS> Find-ByTextManagedObjectCollection -Text $Device1.name

Find a list of managed objects by text

.EXAMPLE
PS> Find-ByTextManagedObjectCollection -Text $Device1.name

Find managed objects which contain the text 'myText' (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z). (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Text,

        # ManagedObject type.
        [Parameter()]
        [string]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory findByText"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Text `
            | Group-ClientRequests `
            | c8y inventory findByText $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Text `
            | Group-ClientRequests `
            | c8y inventory findByText $c8yargs
        }
        
    }

    End {}
}
