﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationBinaryCollection {
<#
.SYNOPSIS
Get application binaries

.DESCRIPTION
A list of all binaries related to the given application will be returned


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_listApplicationBinaries

.EXAMPLE
PS> Get-ApplicationBinaryCollection -Id $App.id

List all of the binaries related to a Hosted (web) application

.EXAMPLE
PS> Get-Application $App.id | Get-ApplicationBinaryCollection

List all of the binaries related to a Hosted (web) application (using pipeline)


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
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications listApplicationBinaries"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customAttachmentCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customBinaryAttachment+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y applications listApplicationBinaries $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications listApplicationBinaries $c8yargs
        }
        
    }

    End {}
}
