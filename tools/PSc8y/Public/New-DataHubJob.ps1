# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DataHubJob {
<#
.SYNOPSIS
Submit a SQL query and retrieve the ID of the Dremio job executing this query

.DESCRIPTION
Submit a SQL query and retrieve the ID of the Dremio job executing this query. The request is asynchronous, i.e., the response does not wait for the query execution to complete.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_jobs_create

.EXAMPLE
PS> New-DataHubJob -Sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms"

Create a new datahub job

.EXAMPLE
PS> New-DataHubJob -Sql "SELECT * FROM alarms" -Context "myTenantIdDataLake", "Dremio", "myTenantId"

Create a new datahub job using context


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The SQL query to execute. The table to query is either referred to with the full path or with the table name if the context defines the path
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Sql,

        # The context in which the query is executed
        [Parameter()]
        [string[]]
        $Context
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub jobs create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Sql `
            | Group-ClientRequests `
            | c8y datahub jobs create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Sql `
            | Group-ClientRequests `
            | c8y datahub jobs create $c8yargs
        }
        
    }

    End {}
}
