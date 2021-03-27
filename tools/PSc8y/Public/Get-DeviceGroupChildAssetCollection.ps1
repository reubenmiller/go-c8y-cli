# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceGroupChildAssetCollection {
<#
.SYNOPSIS
Get child asset collection

.DESCRIPTION
Get a collection of managedObjects child references

.LINK
c8y devicegroups listAssets

.EXAMPLE
PS> DeviceGroupChildAssetCollection -Id $Group.id

Get a list of the child assets of an existing device

.EXAMPLE
PS> DeviceGroupChildAssetCollection -Id $Group.id

Get a list of the child assets of an existing group


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device Group. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups listAssets"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devicegroups listAssets $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devicegroups listAssets $c8yargs
        }
        
    }

    End {}
}
