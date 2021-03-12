# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Tenant {
<#
.SYNOPSIS
Delete tenant

.DESCRIPTION
Delete an existing tenant

.LINK
c8y tenants delete

.EXAMPLE
PS> Remove-Tenant -Id mycompany

Delete a tenant by name (from the management tenant)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y tenants delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y tenants delete $c8yargs
        }
        
    }

    End {}
}
