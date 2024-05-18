# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-RemoteAccessTelnetConfiguration {
<#
.SYNOPSIS
Create telnet configuration

.DESCRIPTION
Create a new Telnet configuration. If no arguments are provided
then sensible defaults will be used.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/remoteaccess_configurations_create-telnet

.EXAMPLE
PS> New-RemoteAccessTelnetConfiguration -Device device01


Create a telnet configuration


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

        # Credentials type
        [Parameter()]
        [string]
        $CredentialsType,

        # Protocol
        [Parameter()]
        [ValidateSet('TELNET')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "remoteaccess configurations create-telnet"
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
            | c8y remoteaccess configurations create-telnet $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations create-telnet $c8yargs
        }
        
    }

    End {}
}
