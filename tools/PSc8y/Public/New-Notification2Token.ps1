# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Notification2Token {
<#
.SYNOPSIS
Create a token

.DESCRIPTION
Create a token to use for subscribing to notifications

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_tokens_create

.EXAMPLE
PS> New-Notification2Token -Name testSubscription -Subscriber testSubscriber -ExpiresInMinutes 30

Create a new token which is valid for 30 minutes


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The subscriber name which the client wishes to be identified with.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Subscriber,

        # The subscription name. This value must match the same that was used when the subscription was created.
        [Parameter()]
        [string]
        $Name,

        # The token expiration duration.
        [Parameter()]
        [long]
        $ExpiresInMinutes,

        # Subscription is shared amongst multiple subscribers
        [Parameter()]
        [ValidateSet('true','false')]
        [string]
        $Shared
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 tokens create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Subscriber `
            | Group-ClientRequests `
            | c8y notification2 tokens create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Subscriber `
            | Group-ClientRequests `
            | c8y notification2 tokens create $c8yargs
        }
        
    }

    End {}
}
