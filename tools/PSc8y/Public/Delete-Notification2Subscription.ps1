# Code generated from specification version 1.0.0: DO NOT EDIT
Function Delete-Notification2Subscription {
<#
.SYNOPSIS
Delete subscription

.DESCRIPTION
Remove a specific subscription by a given ID

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_subscriptions_delete

.EXAMPLE
PS> Delete-Notification2Subscription -Id 12345

Delete a subscription


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y notification2 subscriptions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y notification2 subscriptions delete $c8yargs
        }
        
    }

    End {}
}
