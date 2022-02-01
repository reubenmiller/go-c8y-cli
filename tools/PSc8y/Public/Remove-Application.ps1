# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Application {
<#
.SYNOPSIS
Delete application

.DESCRIPTION
The application can only be removed when its availability is PRIVATE or in other case when it has no subscriptions

Delete an application (by a given ID). This method is not supported by microservice applications.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_delete

.EXAMPLE
PS> Remove-Application -Id $App.id

Delete an application by id

.EXAMPLE
PS> Remove-Application -Id $App.name

Delete an application by name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Force deletion by unsubscribing all tenants from the application first and then deleting the application itself.
        [Parameter()]
        [switch]
        $UnsubscribeAll
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications delete"
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
            | c8y applications delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications delete $c8yargs
        }
        
    }

    End {}
}
