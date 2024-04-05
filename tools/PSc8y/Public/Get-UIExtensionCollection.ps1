# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UIExtensionCollection {
<#
.SYNOPSIS
Get UI extensions collection

.DESCRIPTION
Get a collection of UI extensions by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_extensions_list

.EXAMPLE
PS> Get-UIExtensionCollection -PageSize 100

Get ui extensions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The name of the application.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # The ID of the tenant that owns the applications.
        [Parameter()]
        [string]
        $Owner,

        # The ID of a tenant that is subscribed to the applications but doesn't own them.
        [Parameter()]
        [string]
        $ProvidedFor,

        # The ID of a tenant that is subscribed to the applications.
        [Parameter()]
        [string]
        $Subscriber,

        # The ID of a user that has access to the applications.
        [Parameter()]
        [object[]]
        $User,

        # The ID of a tenant that either owns the application or is subscribed to the applications.
        [Parameter()]
        [string]
        $Tenant,

        # When set to true, the returned result contains applications with an applicationVersions field that is not empty. When set to false, the result will contain applications with an empty applicationVersions field.
        [Parameter()]
        [switch]
        $HasVersions,

        # Application access level for other tenants.
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui extensions list"
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
            | c8y ui extensions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y ui extensions list $c8yargs
        }
        
    }

    End {}
}
