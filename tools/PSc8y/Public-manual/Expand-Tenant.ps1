Function Expand-Tenant {
<#
.SYNOPSIS
Expand the tenants by id or name

.DESCRIPTION
Expand a list of tenants replacing any ids or names with the actual tenant object.

.NOTES
If the given object is already an tenant object, then it is added with no additional lookup

.EXAMPLE
Expand-Tenant "mytenant"

Retrieve the tenant objects by name or id

.EXAMPLE
Get-Tenant *test* | Expand-Tenant

Get all the tenant object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

#>
    [cmdletbinding()]
    Param(
        # List of ids, names or tenant objects
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject
    )

    Process {
        [array] $AllTenants = foreach ($iTenant in $InputObject)
        {
            if ("$iTenant".Contains("*"))
            {
                Get-TenantCollection -PageSize 2000 | Where-Object {
                    $_.id -like $iTenant
                } -WhatIf:$false
            }
            else
            {
                $iTenant
            }
        }

        $AllTenants
    }
}
