Function Get-C8ySessionProperty {
<# 
.SYNOPSIS
Get a property from the current c8y session

.DESCRIPTION
An interface to read properties from the current c8y session, i.e. tenant or host. This is mostly used
internally my other cmdlets in the module to abstract the accessing of such variables in case the environment
variables change in the future (i.e. $env:C8Y_TENANT or $env:C8Y_HOST).

.EXAMPLE
Get-C8ySessionProperty tenant

Get the tenant name of the current c8y cli session

.OUTPUTS
string
#>
    [cmdletbinding()]
    Param(
        # Property name
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [ValidateSet("tenant", "host")]
        [string] $Name
    )

    switch ($Name) {
        "tenant" {
            $env:C8Y_TENANT
        }

        "host" {
            $env:C8Y_HOST
        }
    }
}
