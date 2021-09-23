# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Configuration {
<#
.SYNOPSIS
Create configuration file

.DESCRIPTION
Create a new configuration file (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_create

.EXAMPLE
PS> New-Configuration -Name "agent config" -Description "Default agent configuration" -ConfigurationType "agentConfig" -Data @{$type=@{}}

Create a new configuration file


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # name
        [Parameter()]
        [string]
        $Name,

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
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $DeviceType,

        # File to upload
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $DeviceType `
            | Group-ClientRequests `
            | c8y configuration create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $DeviceType `
            | Group-ClientRequests `
            | c8y configuration create $c8yargs
        }
        
    }

    End {}
}
