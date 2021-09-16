# Code generated from specification version 1.0.0: DO NOT EDIT
Function Install-SoftwareVersion {
<#
.SYNOPSIS
Install software version on a device

.DESCRIPTION
Install software version on a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_install

.EXAMPLE
PS> Get-SoftwareVersion -SoftwareId 12345 -Id $mo.id

Get a software package

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Get-SoftwareVersion

Get a software package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device or agent where the software should be installed
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Software name (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Software,

        # Software version
        [Parameter()]
        [object[]]
        $Version
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions install"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y software versions install $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y software versions install $c8yargs
        }
        
    }

    End {}
}
