# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationCollection {
<#
.SYNOPSIS
Get application collection

.DESCRIPTION
Get a collection of applications by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_list

.EXAMPLE
PS> Get-ApplicationCollection -PageSize 100

Get applications


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application type
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [ValidateSet('APAMA_CEP_RULE','EXTERNAL','HOSTED','MICROSERVICE')]
        [object[]]
        $Type,

        # The name of the application.
        [Parameter()]
        [string]
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
        $HasVersions
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Type `
            | Group-ClientRequests `
            | c8y applications list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Type `
            | Group-ClientRequests `
            | c8y applications list $c8yargs
        }
        
    }

    End {}
}
