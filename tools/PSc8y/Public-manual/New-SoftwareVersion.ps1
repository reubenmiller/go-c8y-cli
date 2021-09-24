# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-SoftwareVersion {
<#
.SYNOPSIS
Create software package version

.DESCRIPTION
Create a new software package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_create

.EXAMPLE
PS> New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}

Create a new version to an existing software package


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Software package id where the version will be added to
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Software,

        # Software package version name, i.e. 1.0.0
        [Parameter()]
        [string]
        $Version,

        # URL to the software package
        [Parameter()]
        [string]
        $Url,

        # File to be uploaded
        [Parameter()]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Software `
            | Group-ClientRequests `
            | c8y software versions create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Software `
            | Group-ClientRequests `
            | c8y software versions create $c8yargs
        }
        
    }

    End {}
}
