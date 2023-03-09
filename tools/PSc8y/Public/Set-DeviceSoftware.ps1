# Code generated from specification version 1.0.0: DO NOT EDIT
Function Set-DeviceSoftware {
<#
.SYNOPSIS
Set/replace software list

.DESCRIPTION
Set/replace a list of software for a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_software_set

.EXAMPLE
PS> Set-DeviceSoftware -Device $software.id -Name ntp -Version 1.0.2 -Type apt

Set/replace the list of software on a device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Software name
        [Parameter()]
        [string]
        $Name,

        # Software version
        [Parameter()]
        [string]
        $Version,

        # Software url
        [Parameter()]
        [string]
        $Url,

        # Software type, e.g. apt
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices software set"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices software set $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices software set $c8yargs
        }
        
    }

    End {}
}
