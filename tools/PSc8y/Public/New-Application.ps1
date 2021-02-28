# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Application {
<#
.SYNOPSIS
Create a new application

.DESCRIPTION
Create a new application using explicit settings

.LINK
c8y applications create

.EXAMPLE
PS> New-Application -Name $AppName -Key "${AppName}-key" -ContextPath $AppName -Type HOSTED

Create a new hosted application


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Name of application (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Name,

        # Shared secret of application (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Key,

        # Type of application. Possible values are EXTERNAL, HOSTED, MICROSERVICE (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('EXTERNAL','HOSTED','MICROSERVICE')]
        [string]
        $Type,

        # Access level for other tenants. Possible values are : MARKET, PRIVATE (default)
        [Parameter()]
        [ValidateSet('MARKET','PRIVATE')]
        [string]
        $Availability,

        # contextPath of the hosted application. Required when application type is HOSTED
        [Parameter()]
        [string]
        $ContextPath,

        # URL to application base directory hosted on an external server. Required when application type is HOSTED
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

        # URL to the external application. Required when application type is EXTERNAL
        [Parameter()]
        [string]
        $ExternalUrl
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = @("")

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            c8y applications create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y applications create $c8yargs
        }
    }

    End {}
}
