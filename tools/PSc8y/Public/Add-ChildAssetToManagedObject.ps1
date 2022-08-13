# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildAssetToManagedObject {
<#
.SYNOPSIS
Assign child asset

.DESCRIPTION
Assigns a group or device to an existing group and marks them as assets

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_assets_assign

.EXAMPLE
PS> Add-ChildAssetToManagedObject -Id $Group1.id -ChildGroup $Group2.id

Create group hierarchy (parent group -> child group)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # New child device to be added to the group as an asset
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $ChildDevice,

        # New child device group to be added to the group as an asset
        [Parameter()]
        [object[]]
        $ChildGroup
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory assets assign"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventory assets assign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventory assets assign $c8yargs
        }
        
    }

    End {}
}
