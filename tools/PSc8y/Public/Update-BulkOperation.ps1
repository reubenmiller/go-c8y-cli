# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-BulkOperation {
<#
.SYNOPSIS
Update bulk operation

.DESCRIPTION
Update bulk operation. Making update on a started bulk operation cancels it and creates/schedules a new one.

.EXAMPLE
PS> Update-BulkOperation -Id $BulkOp.id -CreationRamp 1.5

Update bulk operation wait period between the creation of each operation to 1.5 seconds


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
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
        Get-ClientCommonParameters -Type "Update", "Template" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("CreationRampSec")) {
            $Parameters["creationRampSec"] = $CreationRampSec
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkOperations update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.bulkoperation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Id)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | c8y bulkOperations update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | c8y bulkOperations update $c8yargs
        }
        
    }

    End {}
}
