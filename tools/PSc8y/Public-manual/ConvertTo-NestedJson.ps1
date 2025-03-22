Function ConvertTo-NestedJson {
<#
.SYNOPSIS
Convert object to JSON

.DESCRIPTION
Convert object to JSON but increase the default depth used by ConvertTo-Json

.EXAMPLE
@{example = "one"} | ConvertTo-NestedJson

Convert object to JSON
#>
    [CmdletBinding()]
    param(
        # Input object
        [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline = $true)]
        [AllowNull()]
        [System.Object]
        ${InputObject},

        # Max depth
        [ValidateRange(1, 2147483647)]
        [int]
        ${Depth} = 20,

        # Compress
        [switch]
        ${Compress})

    begin {
        try {
            $outBuffer = $null
            if ($PSBoundParameters.TryGetValue('OutBuffer', [ref]$outBuffer)) {
                $PSBoundParameters['OutBuffer'] = 1
            }

            $wrappedCmd = $ExecutionContext.InvokeCommand.GetCommand('Microsoft.PowerShell.Utility\ConvertTo-Json', [System.Management.Automation.CommandTypes]::Cmdlet)
            $scriptCmd = { & $wrappedCmd @PSBoundParameters }

            $steppablePipeline = $scriptCmd.GetSteppablePipeline($myInvocation.CommandOrigin)
            $steppablePipeline.Begin($PSCmdlet)
        }
        catch {
            throw
        }
    }

    process {
        try {
            $steppablePipeline.Process($_)
        }
        catch {
            throw
        }
    }

    end {
        try {
            $steppablePipeline.End()
        }
        catch {
            throw
        }
    }
}