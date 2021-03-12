# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-AssetFromGroup {
<#
.SYNOPSIS
Unassign asset from group

.DESCRIPTION
Unassign/delete an asset (device or group) from a group

.LINK
c8y inventoryReferences unassignAssetFromGroup

.EXAMPLE
PS> Remove-AssetFromGroup -Group $Group.id -ChildDevice $ChildDevice.id

Unassign a child device from its parent asset


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Asset id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # Child device
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $ChildDevice,

        # Child device group
        [Parameter()]
        [object[]]
        $ChildGroup
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences unassignAssetFromGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences unassignAssetFromGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $ChildDevice `
            | Group-ClientRequests `
            | c8y inventoryReferences unassignAssetFromGroup $c8yargs
        }
        
    }

    End {}
}
