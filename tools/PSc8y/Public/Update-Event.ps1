# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Event {
<#
.SYNOPSIS
Update event

.DESCRIPTION
Update an existing event

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_update

.EXAMPLE
PS> Update-Event -Id $Event.id -Text "example text 1"

Update the text field of an existing event

.EXAMPLE
PS> Update-Event -Id $Event.id -Data @{ my_event = @{ active = $true } }

Update custom properties of an existing event

.EXAMPLE
PS> Get-Event -Id $Event.id | Update-Event -Data @{ my_event = @{ active = $true } }

Update custom properties of an existing event (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Event id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Text description of the event.
        [Parameter()]
        [string]
        $Text
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.event+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y events update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y events update $c8yargs
        }
        
    }

    End {}
}
