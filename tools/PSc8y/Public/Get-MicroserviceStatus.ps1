# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MicroserviceStatus {
<#
.SYNOPSIS
Get microservice status

.DESCRIPTION
Get the status of a microservice which is stored as a managed object


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/microservices_getStatus

.EXAMPLE
PS> Get-MicroserviceStatus -Id 1234 -Dry

Get microservice status

.EXAMPLE
PS> Get-MicroserviceCollection | Get-MicroserviceStatus -Dry

Get microservice status (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Microservice id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices getStatus"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y microservices getStatus $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y microservices getStatus $c8yargs
        }
        
    }

    End {}
}
