# Code generated from specification version 1.0.0: DO NOT EDIT
Function Unregister-Notification2Subscriber {
<#
.SYNOPSIS
Unsubscribe via a token

.DESCRIPTION
Unsubscribe a notification subscriber using the notification token
Once a subscription is made, notifications will be kept until they are consumed by all subscribers who have previously connected to the subscription.

For non-volatile subscriptions, this can result in notifications remaining in storage if never consumed by the application.
They will be deleted if a tenant is deleted. It can take up considerable space in permanent storage for high-frequency notification sources.
Therefore, we recommend you to unsubscribe a subscriber that will never run again.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_tokens_unsubscribe

.EXAMPLE
PS> Unregister-Notification2Subscriber -Token "eyJhbGciOiJSUzI1NiJ9"

Unsubscribe a subscriber using its token


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Subscriptions associated with this token will be removed (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Token
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 tokens unsubscribe"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Token `
            | Group-ClientRequests `
            | c8y notification2 tokens unsubscribe $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Token `
            | Group-ClientRequests `
            | c8y notification2 tokens unsubscribe $c8yargs
        }
        
    }

    End {}
}
