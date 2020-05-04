Function New-Session {
    <#
    .SYNOPSIS
    Get the active Cumulocity Session
    
    .EXAMPLE
    Get-Session
    
    Get the current Cumulocity session
    
    .OUTPUTS
    None
    #>
        [CmdletBinding()]
        Param(
            # Name of the Cumulocity session
            [Parameter(Mandatory = $true)]
            [string]
            $Name,
    
            # Host the type of this alarm, e.g. 'com_cumulocity_events_TamperEvent'. (required)
            [Parameter(Mandatory = $true)]
            [string]
            $Host,
    
            # Tenant the type of this alarm, e.g. 'com_cumulocity_events_TamperEvent'. (required)
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
    
        $args = New-object System.Collections.ArrayList
    
        $null = $args.AddRange(@("sessions", "create"))
    
        if ($Name) {
            $null = $args.AddRange(@("--name", $Name))
        }
        if ($Host) {
            $null = $args.AddRange(@("--host", $Host))
        }
        if ($Tenant) {
            $null = $args.AddRange(@("--tenant", $Tenant))
        }
        if ($Credential.GetNetworkCredential().Username) {
            $null = $args.AddRange(@("--username", $Credential.GetNetworkCredential().Username))
        }
        if ($Credential.GetNetworkCredential().Password) {
            $null = $args.AddRange(@("--password", $Credential.GetNetworkCredential().Password))
        }
        if ($Description) {
            $null = $args.AddRange(@("--description", $Description))
        }
    
        if ($NoTenantPrefix.IsPresent) {
            $null = $args.AddRange("--noTenantPrefix={0}" -f $NoTenantPrefix.ToString().ToLower())
        }
    
        $Path = & $Binary $args
    
        Set-Session -File $Path
    }
    