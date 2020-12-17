Function New-Session {
<#
.SYNOPSIS
Create a new Cumulocity Session

.DESCRIPTION
Create a new Cumulocity session which can be used by the cmdlets. The new session will be automatically activated.

.EXAMPLE
New-Session -Name "develop" -Host "my-tenant-name.eu-latest.cumulocity.com"

Create a new Cumulocity session called develop

.EXAMPLE
New-Session -Host "my-tenant-name.eu-latest.cumulocity.com"

Create a new Cumulocity session. It will prompt for the username and password.

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param(
        # Host url, i.e. https://my-tenant-name.eu-latest.cumulocity.com
        [Parameter(Mandatory = $true)]
        [string]
        $Host,
    
        # Tenant id, i.e. t12345
        [Parameter(Mandatory = $false)]
        [string]
        $Tenant,
    
        # Username
        [Parameter()]
        $Username,

        # Password
        [Parameter()]
        $Password,

        # Name of the Cumulocity session
        [Parameter(Mandatory = $false)]
        [string]
        $Name,
    
        # Description
        [Parameter(Mandatory = $false)]
        [string]
        $Description,
    
        # Don't use tenant name as a prefix to the user name when using Basic Authentication
        [Parameter(Mandatory = $false)]
        [switch]
        $NoTenantPrefix
    )
    
    $Binary = Get-ClientBinary
    
    $c8yargs = New-object System.Collections.ArrayList
    
    $null = $c8yargs.AddRange(@("sessions", "create"))

    if ($Username) {
        $null = $c8yargs.AddRange(@("--username", $Username))
    }

    if ($Password) {
        $null = $c8yargs.AddRange(@("--password", $Password))
    }
    
    if ($Name) {
        $null = $c8yargs.AddRange(@("--name", $Name))
    }
    if ($Host) {
        $null = $c8yargs.AddRange(@("--host", $Host))
    }
    if ($Tenant) {
        $null = $c8yargs.AddRange(@("--tenant", $Tenant))
    }
    if ($Username) {
        $null = $c8yargs.AddRange(@("--username", $Username))
    }
    if ($Password) {
        $null = $c8yargs.AddRange(@("--password", $Password))
    }
    if ($Description) {
        $null = $c8yargs.AddRange(@("--description", $Description))
    }
    
    if ($NoTenantPrefix.IsPresent) {
        $null = $c8yargs.AddRange("--noTenantPrefix={0}" -f $NoTenantPrefix.ToString().ToLower())
    }
    
    $Path = & $Binary $c8yargs
    
    Set-Session -File $Path

    # Test the login
    Invoke-ClientLogin
}
