# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObjectCount {
<#
.SYNOPSIS
Get managed object count

.DESCRIPTION
Retrieve the total number of managed objects (e.g. devices, assets, etc.) registered in your tenant, or a subset based on queries.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_count

.EXAMPLE
PS> Get-ManagedObjectCount

Get count of managed objects


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # List of ids.
        [Parameter()]
        [string[]]
        $Ids,

        # ManagedObject type.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

        # List of managed objects that are owned by the given username.
        [Parameter()]
        [string]
        $Owner,

        # managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z).
        [Parameter()]
        [string]
        $Text,

        # Search for a specific child addition and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAdditionId,

        # Search for a specific child asset and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAssetId,

        # Search for a specific child device and list all the groups to which it belongs.
        [Parameter()]
        [object[]]
        $ChildDeviceId
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory count"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedobjectuser+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Type `
            | Group-ClientRequests `
            | c8y inventory count $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Type `
            | Group-ClientRequests `
            | c8y inventory count $c8yargs
        }
        
    }

    End {}
}
