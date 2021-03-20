# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentApplicationSubscription {
<#
.SYNOPSIS
Get current application subscriptions

.DESCRIPTION
Requires authentication with the application bootstrap user

.LINK
c8y currentapplication listSubscriptions

.EXAMPLE
PS> Get-CurrentApplicationSubscription

List the current application users/subscriptions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currentapplication listSubscriptions"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationUserCollection+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currentapplication listSubscriptions $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currentapplication listSubscriptions $c8yargs
        }
    }

    End {}
}
