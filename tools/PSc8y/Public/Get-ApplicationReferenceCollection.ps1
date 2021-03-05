# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationReferenceCollection {
<#
.SYNOPSIS
Get application reference collection

.DESCRIPTION
Get a collection of application references on a tenant

.LINK
c8y tenants listReferences

.EXAMPLE
PS> Get-ApplicationReferenceCollection -Tenant mycompany

Get a list of referenced applications on a given tenant (from management tenant)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants listReferences"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.applicationReference+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Tenant `
            | Group-ClientRequests `
            | c8y tenants listReferences $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y tenants listReferences $c8yargs
        }
        
    }

    End {}
}
