Function New-HostedApplication {
<#
.SYNOPSIS
New hosted (web) application

.DESCRIPTION
Create a new hosted web application by uploading a zip file which contains a web application

.EXAMPLE
New-HostedApplication -Name $App.id -File "myapp.zip"

Upload application zip file containing the web application

#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # File or Folder of the web application. It should contains a index.html file in the root folder/ or zip file (required)
        [Parameter(Mandatory = $false,
            ValueFromPipeline=$true,
            ValueFromPipelineByPropertyName=$true)]
        [Alias("FullName")]
        [Alias("Path")]
        [string]
        $File,

        # File to be uploaded as a binary
        [Parameter(Mandatory = $false)]
        [string]
        $Name,

        # Shared secret of application. Defaults to the application name with a "-application-key" suffix if not provided.
        [Parameter(Mandatory = $false)]
        [string]
        $Key,

        # contextPath of the hosted application. Defaults to the application name if not provided.
        [Parameter(Mandatory = $false)]
        [string]
        $ContextPath,

        # URL to application base directory hosted on an external server. Required when application type is HOSTED
        [Parameter()]
        [string]
        $ResourcesUrl,

        # Access level for other tenants.  Possible values are : MARKET, PRIVATE (default)
        [Parameter()]
        [ValidateSet('MARKET','PRIVATE')]
        [string]
        $Availability,

        # Don't uploaded the web app binary. Only the application placeholder will be created
        [Parameter()]
        [switch]
        $SkipUpload,

        # Don't subscribe to the application after it has been created and uploaded
        [Parameter()]
        [switch]
        $SkipActivation,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        #
        # Set defaults
        if (!$Key) {
            $Key = $Name
        }

        if (!$ContextPath) {
            $ContextPath = $Name
        }


        $Parameters = @{}

        if ($PSBoundParameters.ContainsKey("Name")) {
            $Parameters["name"] = $Name
        }
        if ($PSBoundParameters.ContainsKey("Availability")) {
            $Parameters["availability"] = $Availability
        }
        if ($PSBoundParameters.ContainsKey("ContextPath")) {
            $Parameters["contextPath"] = $ContextPath
        }
        if ($PSBoundParameters.ContainsKey("ResourcesUrl")) {
            $Parameters["resourcesUrl"] = $ResourcesUrl
        }
        if ($PSBoundParameters.ContainsKey("SkipActivation")) {
            $Parameters["skipActivation"] = $SkipActivation.ToString().ToLower()
        }
        if ($PSBoundParameters.ContainsKey("SkipUpload")) {
            $Parameters["skipUpload"] = $SkipUpload.ToString().ToLower()
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }

    }

    Process {
        # Set empty array if no file was provided (so it still uses the loop, but ignores the file)
        if (!$File) {
            $File = @("")
        }

        foreach ($item in $File) {
            if ($item) {
                $Parameters["file"] = (Resolve-Path $item).ProviderPath
            }

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "applications" `
                -Verb "createHostedApplication" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.application+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
