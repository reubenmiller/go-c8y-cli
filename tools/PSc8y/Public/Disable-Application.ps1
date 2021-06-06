# Code generated from specification version 1.0.0: DO NOT EDIT
Function Disable-Application {
<#
.SYNOPSIS
Unsubscribe application

.DESCRIPTION
Disable/unsubscribe an application from a tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenants_disableApplication

.EXAMPLE
PS> Disable-Application -Tenant t12345 -Application myMicroservice

Disable an application of a tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Application,

        # Tenant id. Defaults to current tenant (based on credentials)
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants disableApplication"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Application `
            | Group-ClientRequests `
            | c8y tenants disableApplication $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Application `
            | Group-ClientRequests `
            | c8y tenants disableApplication $c8yargs
        }
        
    }

    End {}
}
