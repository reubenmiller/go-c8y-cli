# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AgentCollection {
<#
.SYNOPSIS
Get agent collection

.DESCRIPTION
Get a collection of agents based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_list

.EXAMPLE
PS> Get-AgentCollection -Name "sensor*" -Type myType

Get a collection of agents with type "myType", and their names start with "sensor"


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customAgent+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y agents list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y agents list $c8yargs
        }
    }

    End {}
}
