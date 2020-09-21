Function New-Session {
<#
.SYNOPSIS
Create a new Cumulocity Session

.DESCRIPTION
Create a new Cumulocity session which can be used by the cmdlets. The new session will be automatically activated.

.EXAMPLE
New-Session -Name "develop" -Host "https://my-tenant-name.eu-latest.cumulocity.com" -Tenant "t12345"

Create a new Cumulocity session

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param(
        # Name of the Cumulocity session
        [Parameter(Mandatory = $true)]
        [string]
        $Name,
    
        # Host url, i.e. https://my-tenant-name.eu-latest.cumulocity.com
        [Parameter(Mandatory = $true)]
        [string]
        $Host,
    
        # Tenant id, i.e. t12345
        [Parameter(Mandatory = $true)]
        [string]
        $Tenant,
    
        # Credential
        [Parameter(Mandatory = $false, ParameterSetName = 'manual')]
        [ValidateNotNull()]
        [System.Management.Automation.PSCredential]
        [System.Management.Automation.Credential()]
        $Credential = [System.Management.Automation.PSCredential]::Empty,
    
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
    
    if (!$Credential -or ($Credential -eq [System.Management.Automation.PSCredential]::Empty)) {
        $Credential = Get-Credential -Message "Enter the API credentials for the $Tenant C8Y Tenant    (leave-out the the tenant prefix)" -ErrorAction SilentlyContinue
    }
    
    if (!$Credential.UserName -or
        !$Credential.GetNetworkCredential().Password) {
        Write-Warning "Credentials are required to create a Cumulocity session"
        return
    }
    
    $c8yargs = New-object System.Collections.ArrayList
    
    $null = $c8yargs.AddRange(@("sessions", "create"))
    
    if ($Name) {
        $null = $c8yargs.AddRange(@("--name", $Name))
    }
    if ($Host) {
        $null = $c8yargs.AddRange(@("--host", $Host))
    }
    if ($Tenant) {
        $null = $c8yargs.AddRange(@("--tenant", $Tenant))
    }
    if ($Credential.GetNetworkCredential().Username) {
        $null = $c8yargs.AddRange(@("--username", $Credential.GetNetworkCredential().Username))
    }
    if ($Credential.GetNetworkCredential().Password) {
        $null = $c8yargs.AddRange(@("--password", $Credential.GetNetworkCredential().Password))
    }
    if ($Description) {
        $null = $c8yargs.AddRange(@("--description", $Description))
    }
    
    if ($NoTenantPrefix.IsPresent) {
        $null = $c8yargs.AddRange("--noTenantPrefix={0}" -f $NoTenantPrefix.ToString().ToLower())
    }
    
    $Path = & $Binary $c8yargs
    
    Set-Session -File $Path
}
