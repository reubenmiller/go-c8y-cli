# Code generated from specification version 1.0.0: DO NOT EDIT
Function Invoke-UserLogout {
<#
.SYNOPSIS
Logout current user

.DESCRIPTION
Logout the current user. This will invalidate the token associated with the user when using OAUTH_INTERNAL

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/currentuser_logout

.EXAMPLE
PS> Invoke-UserLogout -Dry

Log out the current user


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currentuser logout"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currentuser logout $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currentuser logout $c8yargs
        }
    }

    End {}
}
