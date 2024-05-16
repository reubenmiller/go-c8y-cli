# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-RemoteAccessPassthroughConfiguration {
<#
.SYNOPSIS
Create passthrough configuration

.DESCRIPTION
Create a passthrough configuration which enables you to connect
directly to the device (via Cumulocity IoT) using a native client such as ssh.

After a passthrough connection has been added, you can open a proxy to it using
one of the following commands:

  * c8y remoteaccess server
  * c8y remoteaccess connect ssh


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/remoteaccess_configurations_create-passthrough

.EXAMPLE
PS> New-RemoteAccessPassthroughConfiguration -Device device01


Create a SSH passthrough configuration to the localhost

.EXAMPLE
PS> New-RemoteAccessPassthroughConfiguration -Device device01 -Hostname customhost -Port 1234 -Name "My custom configuration"


Create a SSH passthrough configuration with custom details


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

        # Connection name
        [Parameter()]
        [string]
        $Name,

        # Hostname
        [Parameter()]
        [string]
        $Hostname,

        # Port
        [Parameter()]
        [long]
        $Port,

        # Protocol
        [Parameter()]
        [ValidateSet('PASSTHROUGH')]
        [string]
        $Protocol
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "remoteaccess configurations create-passthrough"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations create-passthrough $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations create-passthrough $c8yargs
        }
        
    }

    End {}
}
