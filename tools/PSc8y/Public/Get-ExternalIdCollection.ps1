# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ExternalIdCollection {
<#
.SYNOPSIS
Get external id collection

.DESCRIPTION
Get a collection of external ids related to an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/identity_list

.EXAMPLE
PS> Get-ExternalIdCollection -Device $Device.id

Get a list of external ids


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "identity list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.externalIdCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.externalId+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y identity list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y identity list $c8yargs
        }
        
    }

    End {}
}
