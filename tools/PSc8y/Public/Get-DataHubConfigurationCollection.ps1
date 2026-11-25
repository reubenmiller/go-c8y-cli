# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubConfigurationCollection {
<#
.SYNOPSIS
Get offloading configurations

.DESCRIPTION
Get offloading configurations

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_configuration_list

.EXAMPLE
PS> Get-DataHubConfigurationCollection

List the datahub offloading configurations


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Last max reported UUID
        [Parameter()]
        [string]
        $LastMaxReportedUUID,

        # Locale
        [Parameter()]
        [string]
        $Locale
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub configuration list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y datahub configuration list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y datahub configuration list $c8yargs
        }
    }

    End {}
}
