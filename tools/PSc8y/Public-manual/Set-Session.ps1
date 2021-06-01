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

.NOTES
On MacOS, you need to hold "shift"+Arrow keys to navigate the list of sessions. Otherwise the VIM style "j" (down) and "k" (up) keys can be also used for navigation

.LINK
c8y sessions set

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
        # SessionFilter list of sessions. Multiple search terms can be provided. A string "Contains" operation
        # is done to match any of the session fields (except password)
        [Parameter(
            ParameterSetName = "ByInteraction",
            Position = 0
        )]
        [string[]] $SessionFilter,

        # Session
        [Parameter(
            ParameterSetName = "ByFile",
            Position = 0
        )]
        [string] $Session
    )

    Process {
        $c8yargs = New-Object System.Collections.ArrayList
        if ($SessionFilter -gt 0) {
            $SearchTerms = $SessionFilter -join " "
            $null = $c8yargs.AddRange(@("--sessionFilter", "$SearchTerms"))
        }

        if ($Session) {
            [void] $c8yargs.AddRange(@("--session", $Session))
        }

        $envvars = c8y sessions set --noColor=false $c8yargs
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "User cancelled set-session. Current session was not changed"
            return
        }

        $envvars | Out-String | Invoke-Expression
    }
}
