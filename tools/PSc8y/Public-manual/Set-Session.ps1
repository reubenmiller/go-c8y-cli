Function Set-Session {
<#
.SYNOPSIS
Set/activate a Cumulocity Session.

.DESCRIPTION
By default the user will be prompted to select from Cumulocity sessions found in their home folder under .cumulocity

.EXAMPLE
Set-Session

.OUTPUTS
String
#>
    [CmdletBinding(
        DefaultParameterSetName = "None"
    )]
    Param(
        # File containing the Cumulocity session data
        [Parameter(Mandatory=$false,
                   Position = 0,
                   ParameterSetName = "ByFile",
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [Alias("FullName")]
        [string] $File,

        # Allow loading Cumulocity session setting from environment variables
        [switch] $UseEnvironment
    )

    Process {

        switch ($PSCmdlet.ParameterSetName) {
            "ByFile" {
                $Path = $File
            }

            default {
                $Binary = Get-ClientBinary
                $args = New-Object System.Collections.ArrayList
                $null = $args.AddRange(@("sessions", "list"))

                if ($UseEnvironment) {
                    $null = $args.Add("--useEnv")
                }
                $Path = & $Binary $args
            }
        }

        # Format path
        $Path = Resolve-Path $Path -ErrorAction SilentlyContinue

        if (!$Path -or !(Test-Path $Path)) {
            Write-Warning "Invalid path"
            return
        }

        Write-Verbose "Setting new session: $Path"
        $env:C8Y_SESSION = Resolve-Path $Path

        # Update environment variables
        Set-EnvironmentVariablesFromSession

        Get-Session
    }
}
