# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubTenant {
<#
.SYNOPSIS
Get tenant configuration

.DESCRIPTION
Get tenant configuration

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_tenant_get

.EXAMPLE
PS> Get-DataHubTenant

Get datahub tenant configuration


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub tenant get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y datahub tenant get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y datahub tenant get $c8yargs
        }
    }

    End {}
}
