# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataBrokerConnectorCollection {
<#
.SYNOPSIS
Get data broker collection

.DESCRIPTION
Get a collection of existing data broker connectors

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/databroker_list

.EXAMPLE
PS> Get-DataBrokerConnectorCollection

Get a list of data broker connectors


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "databroker list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.databrokerConnectorCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.databrokerConnector+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y databroker list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y databroker list $c8yargs
        }
    }

    End {}
}
