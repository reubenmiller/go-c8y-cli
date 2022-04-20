﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-ApplicationBinary {
<#
.SYNOPSIS
New application binary

.DESCRIPTION
For the applications of type microservice and web application to be available for Cumulocity platform users, a binary zip file must be uploaded.

For the microservice application, the zip file must consist of
    * cumulocity.json - file describing the deployment
    * image.tar - executable docker image

For the web application, the zip file must include index.html in the root directory.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_createBinary

.EXAMPLE
PS> New-ApplicationBinary -Id $App.id -File $MicroserviceZip

Upload application microservice binary


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

        # File to be uploaded as a binary (required)
        [Parameter(Mandatory = $true)]
        [string]
        $File
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications createBinary"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y applications createBinary $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications createBinary $c8yargs
        }
        
    }

    End {}
}
