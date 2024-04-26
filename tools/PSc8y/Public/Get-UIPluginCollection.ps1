# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UIPluginCollection {
<#
.SYNOPSIS
Get UI plugin collection

.DESCRIPTION
Get a collection of UI plugins by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_list

.EXAMPLE
PS> Get-UIPluginCollection -PageSize 100

Get UI plugins


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The name of the plugin.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # The ID of the tenant that owns the plugin.
        [Parameter()]
        [string]
        $Owner,

        # The ID of a tenant that is subscribed to the plugin but doesn't own them.
        [Parameter()]
        [string]
        $ProvidedFor,

        # The ID of a tenant that is subscribed to the plugin.
        [Parameter()]
        [string]
        $Subscriber,

        # The ID of a user that has access to the plugin.
        [Parameter()]
        [object[]]
        $User,

        # The ID of a tenant that either owns the plugin or is subscribed to the plugins.
        [Parameter()]
        [string]
        $Tenant,

        # When set to true, the returned result contains plugins with an applicationVersions field that is not empty. When set to false, the result will contain applications with an empty applicationVersions field.
        [Parameter()]
        [switch]
        $HasVersions,

        # Plugin access level for other tenants.
        [Parameter()]
        [ValidateSet('SHARED','PRIVATE','MARKET')]
        [string]
        $Availability
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y ui plugins list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y ui plugins list $c8yargs
        }
        
    }

    End {}
}
