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
PS> New-Tenant -Name "mycompany" -Domain "mycompany" -AdminEmail "admin@example.com" -AdminName "admin" -AdminPass "mys3curep9d8"

Create a new tenant (from the management tenant)

.EXAMPLE
PS> New-Tenant -Name "mycompany" -Domain "mycompany" -AdminEmail "admin@example.com" -AdminName "admin" -SendPasswordResetEmail

Create a new tenant and send a password reset email (from the management tenant)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Company name. Maximum 256 characters
        [Parameter()]
        [string]
        $Name,

        # Domain name to be used for the tenant. Maximum 256 characters
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Domain,

        # Email address of the tenant's administrator
        [Parameter()]
        [string]
        $AdminEmail,

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
        $TenantId,

        # Allow the tenant to create sub-tenants
        [Parameter()]
        [switch]
        $AllowCreateTenants,

        # Send password reset email to the user instead of setting a password
        [Parameter()]
        [switch]
        $SendPasswordResetEmail
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
