# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Notification2SubscriptionCollection {
<#
.SYNOPSIS
Get subscription collection

.DESCRIPTION
Retrieve all subscriptions on your tenant, or a specific subset based on queries.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_subscriptions_list

.EXAMPLE
PS> Get-Notification2SubscriptionCollection

Get existing subscriptions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The managed object ID to which the subscription is associated.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # The context to which the subscription is associated.
        [Parameter()]
        [ValidateSet('mo','tenant')]
        [string]
        $Context
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.subscriptioncollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.subscription+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions list $c8yargs
        }
        
    }

    End {}
}
