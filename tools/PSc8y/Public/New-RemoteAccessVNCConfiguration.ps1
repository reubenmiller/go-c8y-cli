# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-RemoteAccessVNCConfiguration {
<#
.SYNOPSIS
Create vnc configuration

.DESCRIPTION
Create a new VNC configuration. If no arguments are provided
then sensible defaults will be used.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/remoteaccess_configurations_create-vnc

.EXAMPLE
PS> New-RemoteAccessVNCConfiguration -Device device01


Create a VNC configuration that does not require a password

.EXAMPLE
PS> New-RemoteAccessVNCConfiguration -Device device01 -Password 'asd08dcj23dsf{@#9}'

Create a VNC configuration that requires a password


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

        # VNC Password
        [Parameter()]
        [string]
        $Password,

        # Protocol
        [Parameter()]
        [ValidateSet('VNC')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "remoteaccess configurations create-vnc"
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
            | c8y remoteaccess configurations create-vnc $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations create-vnc $c8yargs
        }
        
    }

    End {}
}
