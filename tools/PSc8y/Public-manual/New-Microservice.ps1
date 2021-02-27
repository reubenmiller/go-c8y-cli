Function New-Microservice {
<#
.SYNOPSIS
New microservice

.DESCRIPTION
Create a new microservice or upload a new microservice binary to an already running microservice. By default the microservice will
also be subscribed to/enabled.

The zip file needs to follow the Cumulocity Microservice format.

This cmdlet has several operations

.NOTES
This cmdlet does not support template variables

.EXAMPLE
PS> New-Microservice -File "myapp.zip"

Upload microservice binary. The name of the microservice will be named after the zip file name (without the extension)

If the microservice already exists, then the only the microservice binary will be updated.

.EXAMPLE
PS> New-Microservice -Name "myapp" -File "myapp.zip"

Upload microservice binary with a custom name. Note: If the microservice already exists in the platform

.EXAMPLE
PS> New-Microservice -Name "myapp" -File "./cumulocity.json" -SkipUpload

Create a microservice placeholder named "myapp" for use for local development of a microservice.

The `-File` parameter is provided with the microserivce's manifest file `cumulocity.json` to set the correct required roles of the bootstrap
user which will be automatically created by Cumulocity.

The microservice's bootstrap credentials can be retrieved using `Get-MicroserviceBootstrapUser` cmdlet.

This example is usefuly for local development only, when you want to run the microservice locally (not hosted in Cumulocity).

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # File to be uploaded as a binary (required)
        [Parameter(Mandatory = $true,
        ValueFromPipeline=$true,
        ValueFromPipelineByPropertyName=$true)]
        [string]
        $File,

        # Name of the microservice. An id is also accepted however the name have been previously uploaded.
        [Parameter(Mandatory = $false)]
        [string]
        $Name,

        # Shared secret of application. Defaults to application name if not provided.
        [Parameter()]
        [string]
        $Key,

        # Access level for other tenants.  Possible values are : MARKET, PRIVATE (default)
        [Parameter()]
        [ValidateSet('MARKET','PRIVATE')]
        [string]
        $Availability,

        # ContextPath of the hosted application. Required when application type is HOSTED
        [Parameter()]
        [string]
        $ContextPath,

        # URL to application base directory hosted on an external server. Required when application type is HOSTED
        [Parameter()]
        [string]
        $ResourcesUrl,

        # Skip the uploading of the microservice binary. This is helpful if you want to run the microservice locally
        # and you only need the microservice place holder in order to create microservice bootstrap credentials.
        [Parameter()]
        [switch]
        $SkipUpload,

        # Don't subscribe to the microservice after it has been created and uploaded
        [Parameter()]
        [switch]
        $SkipSubscription
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $Parameters = @{} + $PSBoundParameters
        $Parameters.Remove("File")

        $ArgOptions = @{
            Parameters = $Parameters
            Command = "microservices create"
        }
        $c8yargs = New-ClientArgument @ArgOptions
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }

        foreach ($item in $File) {
            $ic8yArgs = New-Object System.Collections.ArrayList
            if ($item) {
                [void]$ic8yArgs.AddRange(@("--file", (Resolve-Path $item).ProviderPath))
            }
            [void]$ic8yArgs.AddRange($c8yargs)

            if (!$Force -and !$WhatIfPreference) {
                $shouldContinue = $PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )
                if (!$shouldContinue) {
                    continue
                }
            }

            if ($ClientOptions.ConvertToPS) {
                c8y microservices create $ic8yArgs `
                | ConvertFrom-ClientOutput @TypeOptions
            }
            else {
                c8y microservices create $ic8yArgs
            }
        }
    }

    End {}
}
