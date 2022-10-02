# Code generated from specification version 1.0.0: DO NOT EDIT
Function Delete-Notification2SubscriptionBySource {
<#
.SYNOPSIS
Delete subscription by source

.DESCRIPTION
Delete an existing subscription associated to a managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_subscriptions_deleteBySource

.EXAMPLE
PS> Delete-Notification2SubscriptionBySource -Device 12345

Delete a subscription associated with a device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The managed object to which the subscription is associated.
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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions deleteBySource"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions deleteBySource $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions deleteBySource $c8yargs
        }
        
    }

    End {}
}
