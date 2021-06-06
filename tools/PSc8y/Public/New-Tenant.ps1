# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Tenant {
<#
.SYNOPSIS
Create tenant

.DESCRIPTION
Create a new tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenants_create

.EXAMPLE
PS> New-Tenant -Company "mycompany" -Domain "mycompany" -AdminName "admin" -AdminPass "mys3curep9d8"

Create a new tenant (from the management tenant)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Company name. Maximum 256 characters (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Company,

        # Domain name to be used for the tenant. Maximum 256 characters (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Domain,

        # Username of the tenant administrator
        [Parameter()]
        [string]
        $AdminName,

        # Password of the tenant administrator
        [Parameter()]
        [string]
        $AdminPass,

        # A contact name, for example an administrator, of the tenant
        [Parameter()]
        [string]
        $ContactName,

        # An international contact phone number
        [Parameter()]
        [string]
        $ContactPhone,

        # The tenant ID. This should be left bank unless you know what you are doing. Will be auto-generated if not present.
        [Parameter()]
        [string]
        $TenantId
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenant+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Domain `
            | Group-ClientRequests `
            | c8y tenants create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Domain `
            | Group-ClientRequests `
            | c8y tenants create $c8yargs
        }
        
    }

    End {}
}
