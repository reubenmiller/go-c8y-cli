# Code generated from specification version 1.0.0: DO NOT EDIT
Function Install-FirmwareVersion {
<#
.SYNOPSIS
Install firmware version on a device

.DESCRIPTION
Install firmware version on a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_versions_install

.EXAMPLE
PS> Install-FirmwareVersion -Device $mo.id -Firmware linux-iot -Version 1.0.0

Get a firmware version


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device or agent where the firmware should be installed
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Firmware name (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Firmware,

        # Firmware version
        [Parameter()]
        [object[]]
        $Version,

        # Firmware url. TODO, not currently automatically added
        [Parameter()]
        [string]
        $Url
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware versions install"
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
            | c8y firmware versions install $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y firmware versions install $c8yargs
        }
        
    }

    End {}
}
