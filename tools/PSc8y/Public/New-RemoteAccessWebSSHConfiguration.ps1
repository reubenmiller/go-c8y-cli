# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-RemoteAccessWebSSHConfiguration {
<#
.SYNOPSIS
Create web ssh configuration

.DESCRIPTION
Create a new WebSSH configuration. If no arguments are provided
then sensible defaults will be used.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/remoteaccess_configurations_create-webssh

.EXAMPLE
PS> New-RemoteAccessWebSSHConfiguration -Device device01 -Username admin -Password "3Xz7cEj%oAmt#dnUMP*N"


Create a webssh configuration (with username/password authentication)

.EXAMPLE
PS> New-RemoteAccessWebSSHConfiguration -Device device01 -Hostname 127.0.0.1 -Port 2222 -Username admin -PrivateKey "xxxx" -PublicKey "yyyyy"

Create a webssh configuration with a custom hostname and port (with ssh key authentication)


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
        [ValidateSet('USER_PASS','KEY_PAIR','CERTIFICATE')]
        [string]
        $CredentialsType,

        # Private ssh key
        [Parameter()]
        [string]
        $PrivateKey,

        # Public ssh key
        [Parameter()]
        [string]
        $PublicKey,

        # Username
        [Parameter()]
        [string]
        $Username,

        # Username
        [Parameter()]
        [string]
        $Password,

        # Protocol
        [Parameter()]
        [ValidateSet('SSH')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "remoteaccess configurations create-webssh"
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
            | c8y remoteaccess configurations create-webssh $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations create-webssh $c8yargs
        }
        
    }

    End {}
}
