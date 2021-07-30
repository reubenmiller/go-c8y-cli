# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DataBrokerConnector {
<#
.SYNOPSIS
Update data broker

.DESCRIPTION
Update an existing data broker connector

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/databroker_update

.EXAMPLE
PS> Update-DataBroker -Id 12345 -Status SUSPENDED

Change the status of a specific data broker connector by given connector id


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Data broker connector id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # DataBroker status [SUSPENDED].
        [Parameter()]
        [ValidateSet('SUSPENDED')]
        [string]
        $Status
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "databroker update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.databrokerConnector+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y databroker update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y databroker update $c8yargs
        }
        
    }

    End {}
}
