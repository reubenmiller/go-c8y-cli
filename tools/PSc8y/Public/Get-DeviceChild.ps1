# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceChild {
<#
.SYNOPSIS
Get child

.DESCRIPTION
Get managed object child

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_children_get

.EXAMPLE
PS> Get-DeviceChild -Id $Agent.id -Child $Ref.id -ChildType childAdditions

Get an existing child managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Child relationship type (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('childAdditions','childAssets','childDevices')]
        [string]
        $ChildType,

        # Child managed object id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Child
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices children get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices children get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices children get $c8yargs
        }
        
    }

    End {}
}
