# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-DeviceSoftwareCollection {
<#
.SYNOPSIS
Find software

.DESCRIPTION
Find software packages for a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_software_list

.EXAMPLE
PS> Find-DeviceSoftwareCollection -Id 12345

Find all software (from a device)


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

        # Software type
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices software list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customDeviceSoftware+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y devices software list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y devices software list $c8yargs
        }
        
    }

    End {}
}
