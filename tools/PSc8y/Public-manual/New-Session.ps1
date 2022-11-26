Function New-Session {
    <#
.SYNOPSIS
Create a new Cumulocity Session

.DESCRIPTION
Create a new Cumulocity session which can be used by the cmdlets

.LINK
c8y sessions create

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
        $NoTenantPrefix,

        # Allow insecure connection (e.g. when using self-signed certificates)
        [Parameter(Mandatory = $false)]
        [switch]
        $AllowInsecure
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "sessions create"
    }
    
    Process {
        $Path = c8y sessions create $c8yargs
        $code = $LASTEXITCODE

        if ($code -ne 0) {
            Write-Warning "user cancelled create session"
            return
        }

        Write-Host "Created session file. Please use Set-Session to activate it" -ForegroundColor Green
        $Path
    }
}
