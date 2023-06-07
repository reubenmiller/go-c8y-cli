# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Notification2Subscription {
<#
.SYNOPSIS
Create subscription

.DESCRIPTION
Create a subscription

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/notification2_subscriptions_create

.EXAMPLE
PS> New-Notification2Subscription -Name deviceSub -Device 12345 -Context mo -ApiFilter operations

Create a new subscription to operations for a specific device


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

        # The subscription name. Each subscription is identified by a unique name within a specific context.
        [Parameter()]
        [string]
        $Name,

        # The context to which the subscription is associated.
        [Parameter()]
        [ValidateSet('mo','tenant')]
        [string]
        $Context,

        # Transforms the data to only include specified custom fragments. Each custom fragment is identified by a unique name. If nothing is specified here, the data is forwarded as-is.
        [Parameter()]
        [string[]]
        $FragmentsToCopy,

        # Filter notifications by api
        [Parameter()]
        [ValidateSet('alarms','events','managedobjects','measurements','operations','*')]
        [string[]]
        $ApiFilter,

        # The data needs to have the specified value in its type property to meet the filter criteria.
        [Parameter()]
        [string]
        $TypeFilter,

        # Indicates whether the messages for this subscription are persistent or non-persistent, meaning they can be lost if consumer is not connected. >= 1016.x
        [Parameter()]
        [switch]
        $NonPersistent
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "notification2 subscriptions create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.subscriptioncollection+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y notification2 subscriptions create $c8yargs
        }
        
    }

    End {}
}
