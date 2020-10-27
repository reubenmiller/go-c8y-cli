Function Set-Session {
<#
.SYNOPSIS
Set/activate a Cumulocity Session.

.DESCRIPTION
By default the user will be prompted to select from Cumulocity sessions found in their home folder under .cumulocity

Filtering the list is always 

"customer dev" will be split in to two search terms, "customer" and "dev", and only results which contain these two search
terms will be includes in the results. The search is applied to the following fields of the session:

* index
* filename (basename only)
* host
* tenant
* username

.EXAMPLE
Set-Session

Prompt for a list of Cumulocity sessions to select from

.EXAMPLE
Set-Session customer

Set a session interactively but only include sessions where the details contain "customer" in any of the fields

.EXAMPLE
Set-Session customer, dev

Set a session interactively but only includes session where the details includes "customer" and "dev" in any of the fields

.OUTPUTS
String
#>
    [CmdletBinding(
        DefaultParameterSetName = "ByInteraction"
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

        # Filter list of sessions. Multiple search terms can be provided. A string "Contains" operation
        # is done to match any of the session fields (except password)
        [Parameter(
            ParameterSetName = "ByInteraction",
            Position = 0
        )]
        [string[]] $Filter,

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
                $c8yargs = New-Object System.Collections.ArrayList
                $null = $c8yargs.AddRange(@("sessions", "list"))

                if ($Filter -gt 0) {
                    $SearchTerms = $Filter -join " "
                    $null = $c8yargs.AddRange(@("--sessionFilter", "$SearchTerms"))
                }

                if ($UseEnvironment) {
                    $null = $c8yargs.Add("--useEnv")
                }
                $Path = & $Binary $c8yargs

                if ($LASTEXITCODE -ne 0) {
                    Write-Warning "User cancelled set-session. Current session was not changed"
                    return
                }
            }
        }

        if (!$Path -or !(Test-Path $Path)) {
            Write-Warning "Invalid path"
            return
        }

        Write-Verbose "Setting new session: $Path"
        $env:C8Y_SESSION = Resolve-Path $Path

        # Update environment variables
        Set-EnvironmentVariablesFromSession

        # Get OAuth2 and test client authentication
        $null = Invoke-ClientLogin

        if ($LASTEXITCODE -ne 0) {
            Write-Error "$resp"
            return
        }

        Get-Session
    }
}
