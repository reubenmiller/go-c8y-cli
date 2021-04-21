# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Microservice {
<#
.SYNOPSIS
Update microservice details

.DESCRIPTION
Update details of an existing microservice, i.e. availability, context path etc.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/microservices_update

.EXAMPLE
PS> Update-Microservice -Id $App.id -Availability "MARKET"

Update microservice availability to MARKET


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
        $Id,

        # Shared secret of microservice
        [Parameter()]
        [string]
        $Key,

        # Access level for other tenants. Possible values are : MARKET, PRIVATE (default)
        [Parameter()]
        [ValidateSet('MARKET','PRIVATE')]
        [string]
        $Availability,

        # contextPath of the hosted application
        [Parameter()]
        [string]
        $ContextPath,

        # URL to microservice base directory hosted on an external server
        [Parameter()]
        [string]
        $ResourcesUrl
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y microservices update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y microservices update $c8yargs
        }
        
    }

    End {}
}
