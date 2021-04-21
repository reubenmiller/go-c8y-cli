# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-EventBinary {
<#
.SYNOPSIS
Update event binary

.DESCRIPTION
Update an existing event binary


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_updateBinary

.EXAMPLE
PS> Update-EventBinary -Id $Event.id -File $TestFile

Update a binary related to an event


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

        # File to be uploaded as a binary (required)
        [Parameter(Mandatory = $true)]
        [string]
        $File
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events updateBinary"
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
            | c8y events updateBinary $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y events updateBinary $c8yargs
        }
        
    }

    End {}
}
