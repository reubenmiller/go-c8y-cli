# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubJob {
<#
.SYNOPSIS
Retrieve the status of a query job

.DESCRIPTION
Retrieve the status of a query given the ID of the Dremio job executing the query

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_jobs_get

.EXAMPLE
PS> Get-DataHubJob -Id "22feee74-875a-561c-5508-04114bdda000"

Retrieve a datahub job


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The unique identifier of a Dremio job (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub jobs get"
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
            | c8y datahub jobs get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y datahub jobs get $c8yargs
        }
        
    }

    End {}
}
