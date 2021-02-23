# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-BulkOperationCollection {
<#
.SYNOPSIS
Get a collection of bulk operations

.DESCRIPTION
Get a collection of bulk operations

.EXAMPLE
PS> Get-BulkOperationCollection

Get a list of bulk operations


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Include CANCELLED bulk operations
        [Parameter()]
        [switch]
        $WithDeleted
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkOperations list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.bulkOperationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.bulkoperation+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            ,(c8y bulkOperations list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions)
        }
        else {
            c8y bulkOperations list $c8yargs
        }
    }

    End {}
}
