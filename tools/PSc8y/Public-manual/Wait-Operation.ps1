Function Wait-Operation {
<#
.SYNOPSIS
Wait for an operation to be completed (i.e. either in the SUCCESS or FAILED status)

.DESCRIPTION
Wait for an operation to be completed with support for a timeout. Useful when writing scripts
which should only proceed once the operation has finished executing.

.PARAMETER Id
Operation id or object to wait for

.PARAMETER Timeout
Timeout in seconds. Defaults to 30 seconds. i.e. how long should it wait for the operation to be processed

.EXAMPLE
Wait-Operation 1234567

Wait for the operation id

.EXAMPLE
Wait-Operation 1234567 -Timeout 30

Wait for the operation id, and timeout after 30 seconds
#>
    Param(
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Id,

        # Timeout in seconds
        [Alias("TimeoutSec")]
        [double] $Timeout = 30
    )
    Process {
        $ExpirationDate = (Get-Date).AddSeconds($Timeout)

        do {
            $op = Get-Operation -Id $id -AsPSObject

            if ($null -eq $op) {
                # Cancel early if the operation does not exist
                Write-Warning "Could not find operation"
                return;
            }
            Start-Sleep -Milliseconds 200
            $HasExpired = (Get-Date) -ge $ExpirationDate
        } while (!$HasExpired -and $op.status -notmatch "(FAILED|SUCCESSFUL)" -and $op.id)

        if ($HasExpired) {
            Write-Warning "Timeout: Operation is still being processed after $Timeout seconds. Operation: $id"
            $op
            return
        }

        switch ($op.status) {
            "FAILED" {
                Write-Warning ("Operation failed [id={0}]. Reason: {1}" -f $op.id, $op.failureReason)
                break;
            }
            "SUCCESSFUL" {
                Write-Verbose "Operation was successful"
                break;
            }
            default {
                throw "Unknown operation status. $($op.status)"
            }
        }
        $op
    }
}
