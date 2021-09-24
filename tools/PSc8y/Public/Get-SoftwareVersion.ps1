# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SoftwareVersion {
<#
.SYNOPSIS
Get software package version

.DESCRIPTION
Get an existing software package version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_get

.EXAMPLE
PS> Get-SoftwareVersion -Software 12345 -Id $mo.id

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
        # Software Package version id or name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Software package id (used to help completion be more accurate)
        [Parameter()]
        [object[]]
        $Software,

        # Include parent references
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions get"
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
            | c8y software versions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y software versions get $c8yargs
        }
        
    }

    End {}
}
