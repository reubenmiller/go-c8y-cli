# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentApplication {
<#
.SYNOPSIS
Get current application

.DESCRIPTION
Getting the current application only works when using bootstrap credentials from an application (not user credentials)


.LINK
c8y currentApplication get

.EXAMPLE
PS> Get-CurrentApplication

Get the current application (requires using application credentials)


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currentApplication get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currentApplication get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currentApplication get $c8yargs
        }
    }

    End {}
}
