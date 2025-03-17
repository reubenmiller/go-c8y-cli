# Code generated from specification version 1.0.0: DO NOT EDIT
Function Register-Device {
<#
.SYNOPSIS
Register device with username/password and manual device approval/bootstrapping

.DESCRIPTION
Register a new device (request) where the device is using the manual device bootstrapping
process to retrieve its device credentials (username/password).

See Cumulocity docs for more details: https://cumulocity.com/docs/2024/device-integration/rest/


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceregistration_register

.EXAMPLE
PS> Register-Device -Id "ASDF098SD1J10912UD92JDLCNCU8"

Register a new device

.EXAMPLE
PS> Register-Device -Id "ASDF098SD1J10912UD92JDLCNCU8" -Group "My Group"

Register a new device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device identifier. Max: 1000 characters. E.g. IMEI (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Type of the device
        [Parameter()]
        [string]
        $Type,

        # Group to which the device will be assigned
        [Parameter()]
        [object[]]
        $Group
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceregistration register"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.newDeviceRequest+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration register $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceregistration register $c8yargs
        }
        
    }

    End {}
}
