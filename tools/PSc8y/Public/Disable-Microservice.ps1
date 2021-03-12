# Code generated from specification version 1.0.0: DO NOT EDIT
Function Disable-Microservice {
<#
.SYNOPSIS
unsubscribe microservice

.DESCRIPTION
Disable (unsubscribe) a microservice from the current tenant


.LINK
c8y microservices disable

.EXAMPLE
PS> Disable-Microservice -Id $App.id

Disable (unsubscribe) to a microservice


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

        # Tenant id
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices disable"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y microservices disable $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y microservices disable $c8yargs
        }
        
    }

    End {}
}
