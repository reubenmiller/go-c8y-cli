# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Microservice {
<#
.SYNOPSIS
Delete microservice

.DESCRIPTION
Info: The application can only be removed when its availability is PRIVATE or in other case when it has no subscriptions.

.LINK
c8y microservices delete

.EXAMPLE
PS> Remove-Microservice -Id $App.id

Delete a microservice by id

.EXAMPLE
PS> Remove-Microservice -Id $App.name

Delete a microservice by name


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices delete"
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
            | c8y microservices delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y microservices delete $c8yargs
        }
        
    }

    End {}
}
