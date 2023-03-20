# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubQueryResult {
<#
.SYNOPSIS
Execute a SQL query and retrieve the results

.DESCRIPTION
Execute a SQL query and retrieve the results

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_query

.EXAMPLE
PS> Get-DataHubQueryResult -Sql "SELECT * FROM myTenantIdDataLake.Dremio.myTenantId.alarms"

Get a list of alarms from datahub


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The version of the high-performance API
        [Parameter()]
        [string]
        $Version,

        # The SQL query to execute
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Sql,

        # The maximum number of query results
        [Parameter()]
        [long]
        $Limit,

        # The response format, which is either DREMIO or PANDAS. The DREMIO format is the same response format as provided by the sql endpoint of the Standard API. The PANDAS format fits to the data format the Pandas library for Python expects.
        [Parameter()]
        [ValidateSet('DREMIO','PANDAS')]
        [string]
        $Format
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub query"
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
            | c8y datahub query $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Sql `
            | Group-ClientRequests `
            | c8y datahub query $c8yargs
        }
        
    }

    End {}
}
