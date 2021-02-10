Function Get-AssetParent {
<# 
.SYNOPSIS
Get asset parent references for a asset

.DESCRIPTION
Get the parent of an existing assert by using the references. The cmdlet supports returning
various forms of the parent references, i.e. immediate parent, parent or the parent, or the
full parental references.

.EXAMPLE
Get-AssetParent asset0*

Get the direct (immediate) parent of the given asset

.EXAMPLE
Get-AssetParent -All

Return an array of parent assets where the first element in the array is the root asset, and the last is the direct parent of the given asset.

.EXAMPLE
Get-AssetParent -RootParent

Returns the root parent. In most cases this will be the agent

#>
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding=$true,
        DefaultParameterSetName = "ByLevel",
        HelpUri='',
        ConfirmImpact = 'None')]
    Param(
        # Asset id, name or object. Wildcards accepted
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0)]
        [object[]] $Asset,


        # Level to navigate backward from the given asset to its parent/s
        # 1 = direct parent
        # 2 = parent of its parent
        # If the Level is too large, then the root parent will be returned
        [Parameter(
            ParameterSetName = "ByLevel",
            Position = 1)]
        [ValidateRange(1,100)]
        [int] $Level = 1,

        # Return the top level / root parent
        [Parameter(
            ParameterSetName = "Root")]
        [switch] $RootParent,

        # Return a list of all parent assets
        [Parameter(
            ParameterSetName = "All")]
        [switch] $All
    )

    Process {
        # Get list of ids
        $Ids = Expand-Id $Asset
        
        $Results = foreach ($iasset in @(Get-ManagedObjectCollection -Ids $Ids -WithParents))
        {
            $Parents = @($iasset.assetParents.references.managedObject | Foreach-Object {
                if ($null -ne $_.id) {
                    New-Object psobject -Property @{
                        id = $_.id;
                        name = $_.name;
                    }
                }
            })

            # Reverse the order because Cumulocity returns the references in order from number of steps from the given asset.
            # So the asset closest to the given asset is first.
            [array]::Reverse($Parents)

            switch ($PSCmdlet.ParameterSetName) {
                "ByLevel" {
                    # Convert to array index (don't need minus 1 because Level is also a 1 based index (same as Count))
                    $Index = $Parents.Count - $Level
                    
                    if ($Index -lt 0) {
                        $Index = 0
                    }
                    if ($Index -ge $Parents.Count) {
                        $Index = $Parents.Count - 1
                    }

                    if ($null -ne $Parents[$Index].id) {
                        $Parents[$Index]
                    }
                }

                "Root" {
                    if ($null -ne $Parents[0].id) {
                        $Parents[0]
                    }
                }

                "All" {
                    $Parents
                }
            }
        }

        $Results `
            | Select-Object `
            | Add-PowershellType -Type "c8y.parentReferences"
    }
}
