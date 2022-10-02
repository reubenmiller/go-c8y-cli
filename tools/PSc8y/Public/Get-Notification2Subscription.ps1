# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Notification2Subscription {
<#
.SYNOPSIS
Get subscription

.DESCRIPTION
Get a subscription by id

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_subscriptions_get

.EXAMPLE
PS> Get-Notification2Subscription -Id 12345

Get an existing subscription


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Unique identifier of the notification subscription. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.subscriptioncollection+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y notification2 subscriptions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y notification2 subscriptions get $c8yargs
        }
        
    }

    End {}
}
