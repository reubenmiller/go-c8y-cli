# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentUser {
<#
.SYNOPSIS
Get user

.DESCRIPTION
Get the user representation associated with the current credentials used by the request

.EXAMPLE
PS> Get-CurrentUser

Get the current user


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get" -BoundParameters $PSBoundParameters
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users getCurrentUser"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.currentUser+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y users getCurrentUser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y users getCurrentUser $c8yargs
        }
    }

    End {}
}
