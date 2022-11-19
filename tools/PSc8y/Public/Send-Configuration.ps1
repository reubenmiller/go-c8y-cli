# Code generated from specification version 1.0.0: DO NOT EDIT
Function Send-Configuration {
<#
.SYNOPSIS
Send configuration to a device via an operation

.DESCRIPTION
Create a new operation to send configuration to an agent or device.

If you provide the reference to the configuration (via id or name), then the configuration's
url and type will be automatically added to the operation.

You may also manually set the url and configurationType rather than looking up the configuration
file in the configuration repository.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_send

.EXAMPLE
PS> Send-Configuration -Device mydevice -Configuration 12345

Send a configuration file to a device

.EXAMPLE
PS> Get-DeviceCollection | Send-Configuration -Configuration 12345

Send a configuration file to multiple devices


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the target device on which this operation should be performed.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Text description of the operation.
        [Parameter()]
        [string]
        $Description,

        # Configuration type
        [Parameter()]
        [string]
        $ConfigurationType,

        # Url to the configuration
        [Parameter()]
        [string]
        $Url,

        # Configuration file (managedObject) id
        [Parameter()]
        [object[]]
        $Configuration
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration send"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.operation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y configuration send $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y configuration send $c8yargs
        }
        
    }

    End {}
}
