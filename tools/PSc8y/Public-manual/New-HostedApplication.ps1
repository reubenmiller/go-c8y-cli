Function New-HostedApplication {
<#
.SYNOPSIS
New hosted (web) application

.DESCRIPTION
Create a new hosted web application by uploading a zip file which contains a web application

.LINK
c8y applications createHostedApplication

.EXAMPLE
New-HostedApplication -Name $App.id -File "myapp.zip"

Upload application zip file containing the web application

#>
    [cmdletbinding(PositionalBinding=$true, HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # File or Folder of the web application. It should contains a index.html file in the root folder/ or zip file
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

        # Don't activate to the application after it has been created and uploaded
        [Parameter()]
        [switch]
        $SkipActivation
    )

    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $Parameters = @{} + $PSBoundParameters
        $Parameters.Remove("File")

        #
        # Set defaults
        if ($Parameters.ContainsKey("Name")) {
            $Parameters["Name"] = $Name
        }

        if (!$Parameters["Key"]) {
            $Parameters["Key"] = $Name
        }

        if (!$Parameters["ContextPath"]) {
            $Parameters["ContextPath"] = $Name
        }

        $ArgOptions = @{
            Parameters = $Parameters
            Command = "applications createHostedApplication"
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
        # Set empty array if no file was provided (so it still uses the loop, but ignores the file)
        if (!$File) {
            $File = @("")
        }

        foreach ($item in $File) {
            $ic8yArgs = New-Object System.Collections.ArrayList
            if ($item) {
                [void]$ic8yArgs.AddRange(@("--file", (Resolve-Path $item).ProviderPath))
            }
            [void]$ic8yArgs.AddRange($c8yargs)

            if ($ClientOptions.ConvertToPS) {
                c8y applications createHostedApplication $ic8yArgs `
                | ConvertFrom-ClientOutput @TypeOptions
            }
            else {
                c8y applications createHostedApplication $ic8yArgs
            }
            
        }
    }

    End {}
}
