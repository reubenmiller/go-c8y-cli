# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantVersion {
<#
.SYNOPSIS
Get tenant version

.DESCRIPTION
Get tenant platform (backend) version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/currenttenant_version

.EXAMPLE
PS> Get-TenantVersion

Get the Cumulocity backend version


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currenttenant version"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currenttenant version $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currenttenant version $c8yargs
        }
    }

    End {}
}
