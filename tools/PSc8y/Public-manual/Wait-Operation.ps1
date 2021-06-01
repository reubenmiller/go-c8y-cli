Function Wait-Operation {
<#
.SYNOPSIS
Wait for an operation to be completed (i.e. either in the SUCCESS or FAILED status)

.DESCRIPTION
Wait for an operation to be completed with support for a timeout. Useful when writing scripts
which should only proceed once the operation has finished executing.

.EXAMPLE
Wait-Operation 1234567

Wait for the operation id

.EXAMPLE
Wait-Operation 1234567 -Duration 30s

Wait for the operation id for a max duration of 30 seconds
#>
    Param(
        # Operation id or object to wait for
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Id,

        # Wait for status
        [string[]] $Status = "SUCCESSFUL",

        # Duration to wait for, i.e. 10s, 1m. Defaults to 30s. i.e. how long should it wait for the operation to be processed
        [string] $Duration
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }
    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations wait"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }
    Process {
        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y operations wait $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y operations wait $c8yargs
        }
    }
}
