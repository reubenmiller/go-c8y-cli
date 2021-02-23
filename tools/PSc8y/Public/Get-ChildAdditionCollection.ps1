# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAdditionCollection {
<#
.SYNOPSIS
Get a collection of managedObjects child additions

.DESCRIPTION
Get a collection of managedObjects child additions

.EXAMPLE
PS> Get-ChildAdditionCollection -Id $software.id

Get a list of the child additions of an existing managed object

.EXAMPLE
PS> Get-ManagedObject -Id $software.id | Get-ChildAdditionCollection

Get a list of the child additions of an existing managed object (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventoryReferences listChildAdditions"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            ,($Id `
            | Group-ClientRequests `
            | c8y inventoryReferences listChildAdditions $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions)
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventoryReferences listChildAdditions $c8yargs
        }
        
    }

    End {}
}
