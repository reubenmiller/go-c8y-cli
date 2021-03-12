# Code generated from specification version 1.0.0: DO NOT EDIT
Function Enable-Application {
<#
.SYNOPSIS
Subscribe application

.DESCRIPTION
Enable/subscribe an application to a tenant

.LINK
c8y tenants enableApplication

.EXAMPLE
PS> Enable-Application -Tenant mycompany -Application myMicroservice

Enable an application of a tenant


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
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants enableApplication"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationReference+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Application `
            | Group-ClientRequests `
            | c8y tenants enableApplication $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Application `
            | Group-ClientRequests `
            | c8y tenants enableApplication $c8yargs
        }
        
    }

    End {}
}
