Function ConvertTo-NestedJson {
<#
.SYNOPSIS
Convert object to nested JSON

.DESCRIPTION
Convert object to nested JSON

.ForwardHelpTargetName Microsoft.PowerShell.Utility\ConvertTo-Json
.ForwardHelpCategory Cmdlet

#>
    [CmdletBinding(HelpUri = 'https://go.microsoft.com/fwlink/?LinkID=2096925', RemotingCapability = 'None')]
    param(
        # Input object to translate to json
        [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline = $true)]
        [AllowNull()]
        [System.Object]
        ${InputObject},

        # Depth. Only serialize until the given depth
        [ValidateRange(1, 2147483647)]
        [int]
        ${Depth} = 20,

        # Use compressed json (not pretty printed)
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