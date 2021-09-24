# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Configuration {
<#
.SYNOPSIS
Update configuration

.DESCRIPTION
Update an existing configuration file (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_update

.EXAMPLE
PS> Update-Configuration -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }

Update a configuration file

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Update-Configuration -Data @{ com_my_props = @{ value = 1 } }

Update a configuration file (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Configuration package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # New configuration name
        [Parameter()]
        [string]
        $NewName,

        # Description of the configuration package
        [Parameter()]
        [string]
        $Description,

        # Configuration type
        [Parameter()]
        [string]
        $ConfigurationType,

        # URL link to the configuration file
        [Parameter()]
        [string]
        $Url,

        # Device type filter. Only allow configuration to be applied to devices of this type
        [Parameter()]
        [string]
        $DeviceType,

        # File to be uploaded
        [Parameter()]
        [string]
        $File
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y configuration update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y configuration update $c8yargs
        }
        
    }

    End {}
}
