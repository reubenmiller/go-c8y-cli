﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-CurrentApplication {
<#
.SYNOPSIS
Update current application

.DESCRIPTION
Required authentication with bootstrap user

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/currentapplication_update

.EXAMPLE
PS> Update-CurrentApplication -Data @{ myCustomProp = @{ value1 = 1}}

Update custom properties of the current application (requires using application credentials)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Name of application
        [Parameter()]
        [string]
        $Name,

        # Shared secret of application
        [Parameter()]
        [string]
        $Key,

        # Application will be applied to this type of documents, possible values [ALARM, AUDIT, EVENT, MEASUREMENT, OPERATION, *].
        [Parameter()]
        [ValidateSet('MARKET','PRIVATE')]
        [string]
        $Availability,

        # contextPath of the hosted application
        [Parameter()]
        [string]
        $ContextPath,

        # URL to application base directory hosted on an external server
        [Parameter()]
        [string]
        $ResourcesUrl,

        # authorization username to access resourcesUrl
        [Parameter()]
        [string]
        $ResourcesUsername,

        # authorization password to access resourcesUrl
        [Parameter()]
        [string]
        $ResourcesPassword,

        # URL to the external application
        [Parameter()]
        [string]
        $ExternalUrl
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currentapplication update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currentapplication update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currentapplication update $c8yargs
        }
    }

    End {}
}
