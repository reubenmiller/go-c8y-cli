# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-AssetToGroup {
<#
.SYNOPSIS
Assign child asset

.DESCRIPTION
Assigns a group or device to an existing group and marks them as assets

.LINK
c8y inventoryReferences createChildAsset

.EXAMPLE
PS> Add-AssetToGroup -Group $Group1.id -NewChildGroup $Group2.id

Create group hierarchy (parent group -> child group)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # New child device to be added to the group as an asset
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChildDevice,

        # New child device group to be added to the group as an asset
        [Parameter()]
        [object[]]
        $NewChildGroup
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences createChildAsset"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $NewChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences createChildAsset $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $NewChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences createChildAsset $c8yargs
        }
        
    }

    End {}
}
