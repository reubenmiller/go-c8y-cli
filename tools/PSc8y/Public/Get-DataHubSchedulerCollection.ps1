# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubSchedulerCollection {
<#
.SYNOPSIS
List scheduler items

.DESCRIPTION
List scheduler items

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_scheduler_list

.EXAMPLE
PS> Get-DataHubSchedulerCollection

List the datahub scheduled items


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Job type
        [Parameter()]
        [ValidateSet('COMPACTION')]
        [string]
        $JobType,

        # Offset
        [Parameter()]
        [long]
        $Offset,

        # Next offset
        [Parameter()]
        [long]
        $NextOffset
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub scheduler list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y datahub scheduler list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y datahub scheduler list $c8yargs
        }
    }

    End {}
}
