# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAssetReference {
<#
.SYNOPSIS
Get child asset reference

.DESCRIPTION
Get managed object child asset reference

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_assets_get

.EXAMPLE
PS> Get-ChildAssetReference -Asset $Agent.id -Reference $Ref.id

Get an existing child asset reference


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Asset id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Asset,

        # Asset reference id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Reference
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory assets get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Asset `
            | Group-ClientRequests `
            | c8y inventory assets get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Asset `
            | Group-ClientRequests `
            | c8y inventory assets get $c8yargs
        }
        
    }

    End {}
}
