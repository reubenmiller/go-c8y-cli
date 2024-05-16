# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RemoteAccessConfigurationCollection {
<#
.SYNOPSIS
List remote access configurations

.DESCRIPTION
List the remote access configurations already configured for a device


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/remoteaccess_configurations_list

.EXAMPLE
PS> Get-RemoteAccessConfigurationCollection -Device device01

List remote access configurations for a given device


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
        $Device
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "remoteaccess configurations list"
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
            | c8y remoteaccess configurations list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y remoteaccess configurations list $c8yargs
        }
        
    }

    End {}
}
