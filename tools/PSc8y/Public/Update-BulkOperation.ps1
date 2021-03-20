# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-BulkOperation {
<#
.SYNOPSIS
Update bulk operation

.DESCRIPTION
Update bulk operation. Making update on a started bulk operation cancels it and creates/schedules a new one.

.LINK
c8y bulkoperations update

.EXAMPLE
PS> Update-BulkOperation -Id $BulkOp.id -CreationRamp 1.5

Update bulk operation wait period between the creation of each operation to 1.5 seconds


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Bulk Operation id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Delay between every operation creation. (required)
        [Parameter(Mandatory = $true)]
        [float]
        $CreationRampSec
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkoperations update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.bulkoperation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y bulkoperations update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y bulkoperations update $c8yargs
        }
        
    }

    End {}
}
