# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MicroserviceCollection {
<#
.SYNOPSIS
Get collection of microservices

.DESCRIPTION
Get a collection of microservices in the current tenant


.EXAMPLE
PS> Get-MicroserviceCollection -PageSize 100

Get microservices


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application type
        [Parameter()]
        [ValidateSet('MICROSERVICE')]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y microservices list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y microservices list $c8yargs
        }
    }

    End {}
}
