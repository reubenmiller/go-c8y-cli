Function Watch-Notification2Subscription {
<#
.SYNOPSIS
Subscribe to a subscription

.DESCRIPTION
Subscribe to an existing subscription. If no token is provided, a token will be created automatically before starting the realtime client

.LINK
https://goc8ycli.netlify.app/docs/cli/c8y/notification2/subscriptions/c8y_notification2_subscriptions_subscribe/

.EXAMPLE
PS> Watch-Notification2Subscription -Name registration

Start listening to a subscription name registration

.EXAMPLE
PS> Watch-Notification2Subscription -Name registration -ActionTypes CREATE -ActionTypes UPDATE

Start listening to a subscription name registration but only include CREATE and UPDATE action types (ignoring DELETE)

.EXAMPLE
PS> Watch-Notification2Subscription -Name registration -Duration 10min

Subscribe to a subscription for 10mins then exit

.EXAMPLE
PS> Watch-Notification2Subscription -Name registration -Token "ey123123123123"

Subscribe using a given token (instead of generating a token)

#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The subscription name. Each subscription is identified by a unique name within a specific context
        [Parameter()]
        [string]
        $Name,

        # Token for the subscription. If not provided, then a token will be created
        [Parameter()]
        [string]
        $Token,

        # The subscriber name which the client wishes to be identified with. Defaults to goc8ycli
        [Parameter()]
        [string]
        $Subscriber,

        # Consumer name. Required for shared subscriptions
        [Parameter()]
        [string]
        $Consumer,

        # Only listen for specific action types, CREATE, UPDATE or DELETE (client side filtering)
        [Parameter()]
        [ValidateSet('CREATE','UPDATE','DELETE')]
        [string[]]
        $ActionTypes,

        # Subscription duration
        [Parameter()]
        [string]
        $Duration,

        # Token expiration duration
        [Parameter()]
        [int]
        $ExpiresInMinutes
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions subscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y notification2 subscriptions subscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y notification2 subscriptions subscribe $c8yargs
        }
        
    }

    End {}
}
