﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Copy-Application {
<#
.SYNOPSIS
Copy application

.DESCRIPTION
A POST request to the 'clone' resource creates a new application based on an already existing one.

The properties are copied to the newly created application. For name, key and context path a 'clone' prefix is added in order to be unique.

If the target application is hosted and has an active version, the new application will have the active version with the same content.

The response contains a representation of the newly created application.

Required role ROLE_APPLICATION_MANAGEMENT_ADMIN


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_copy

.EXAMPLE
PS> Copy-Application -Id $App.id

Copy an existing application


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
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications copy"
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
            | c8y applications copy $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications copy $c8yargs
        }
        
    }

    End {}
}
