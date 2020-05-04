Function Wait-Operation {
<#
.SYNOPSIS
Wait for an operation to be completed (i.e. either in the SUCCESS or FAILED status)

.PARAMETER Id
Operation id or object to wait for

.PARAMETER TimeoutSec
Timeout in seconds. Defaults to 30 seconds. i.e. how long should it wait for the operation to be processed

.EXAMPLE
Wait-Operation 1234567

Wait for the operation id

.EXAMPLE
Wait-Operation 1234567 -TimeoutSec 30

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

        [int] $TimeoutSec = 30
    )
    Process {
        $ExpirationDate = (Get-Date).AddSeconds($TimeoutSec)

        do {
            $op = Get-Operation -Id $id

            if ($null -eq $op) {
                # Cancel early if the operation does not exist
                Write-Warning "Could not find operation"
                return;
            }
            Start-Sleep -Milliseconds 200
            $HasExpired = (Get-Date) -ge $ExpirationDate
        } while (!$HasExpired -and $op.status -notmatch "(FAILED|SUCCESSFUL)" -and $op.id)

        if ($HasExpired) {
            Write-Warning "Timeout: Operation is still being processed after $TimeoutSec seconds. Operation: $id"
            $op
            return
        }

        switch ($op.status) {
            "FAILED" {
                Write-Warning ("Operation failed [id={1}]. Reason: {0}" -f $op.id, $op.failureReason)
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
