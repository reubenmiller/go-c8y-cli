# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAssetCollection {
<#
.SYNOPSIS
Get child asset collection

.DESCRIPTION
Get a collection of managedObjects child references

.LINK
c8y inventoryReferences listChildAssets

.EXAMPLE
PS> Get-ChildAssetCollection -Group $Group.id

Get a list of the child assets of an existing device

.EXAMPLE
PS> Get-ChildAssetCollection -Group $Group.id

Get a list of the child assets of an existing group


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device.
        [Parameter()]
        [object[]]
        $Device,

        # Group.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Group
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences listChildAssets"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Group `
            | Group-ClientRequests `
            | c8y inventoryReferences listChildAssets $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Group `
            | Group-ClientRequests `
            | c8y inventoryReferences listChildAssets $c8yargs
        }
        
    }

    End {}
}
