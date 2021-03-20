# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantOptionEditable {
<#
.SYNOPSIS
Update tenant option edit setting

.DESCRIPTION
Update read-only setting of an existing tenant option Required role:: ROLE_OPTION_MANAGEMENT_ADMIN, Required tenant management Example Request:: Update access.control.allow.origin option.


.LINK
c8y tenantoptions updateEdit

.EXAMPLE
PS> Update-TenantOptionEditable -Category "c8y_cli_tests" -Key "$option8" -Editable "true"

Update editable property for an existing tenant option


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant Option category (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Category,

        # Tenant Option key (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key,

        # Whether the tenant option should be editable or not (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('true','false')]
        [string]
        $Editable
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantoptions updateEdit"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y tenantoptions updateEdit $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y tenantoptions updateEdit $c8yargs
        }
        
    }

    End {}
}
