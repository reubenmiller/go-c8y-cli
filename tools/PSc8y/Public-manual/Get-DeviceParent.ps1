Function Get-DeviceParent {
<# 
.SYNOPSIS
Get device parent references for a device

.EXAMPLE
Get-DeviceParent device0*

Get the direct (immediate) parent of the given device

.EXAMPLE
Get-DeviceParent -All

Return an array of parent devices where the first element in the array is the root device, and the last is the direct parent of the given device.

.EXAMPLE
Get-DeviceParent -RootParent

Returns the root parent. In most cases this will be the agent

#>
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding=$true,
        DefaultParameterSetName = "ByLevel",
        HelpUri='',
        ConfirmImpact = 'None')]
    Param(
        # Device id, name or object. Wildcards accepted
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0)]
        [object[]] $Device,


        # Level to navigate backward from the given device to its parent/s
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

        # Return a list of all parent devices
        [Parameter(
            ParameterSetName = "All")]
        [switch] $All
    )

    Process {
        # Get list of ids
        $Ids = (Expand-Device $Device) | Select-Object -ExpandProperty id
        
        $Results = foreach ($iDevice in @(Get-ManagedObjectCollection -Device $Ids -WithParents))
        {
            $Parents = @($iDevice.deviceParents.references.managedObject | Foreach-Object {
                if ($null -ne $_.id) {
                    New-Object psobject -Property @{
                        id = $_.id;
                        name = $_.name;
                    }
                }
            })

            # Reverse the order because Cumulocity returns the references in order from number of steps from the given device.
            # So the device closest to the given device is first.
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
            | Expand-Device `
            | Select-Object `
            | Add-PowershellType -Type "c8y.parentReferences"
    }
}
