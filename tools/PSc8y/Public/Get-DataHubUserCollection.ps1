# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DataHubUserCollection {
<#
.SYNOPSIS
List the data hub users

.DESCRIPTION
List the data hub users

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/datahub_users_list

.EXAMPLE
PS> Get-DataHubUserCollection

List the datahub users


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "datahub users list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y datahub users list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y datahub users list $c8yargs
        }
    }

    End {}
}
