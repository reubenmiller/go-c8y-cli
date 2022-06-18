# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceCertificate {
<#
.SYNOPSIS
Update trusted device certificate

.DESCRIPTION
Update settings of an existing trusted device certificate

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificates_update

.EXAMPLE
PS> Update-DeviceCertificate -Id abcedef0123456789abcedef0123456789 -Status DISABLED

Update device certificate by id/fingerprint

.EXAMPLE
PS> Update-DeviceCertificate -Id "MyCert" -Status DISABLED

Update device certificate by name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Certificate fingerprint or name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Tenant id (required)
        [Parameter(Mandatory = $true)]
        [object]
        $Tenant,

        # Certificate name
        [Parameter()]
        [string]
        $Name,

        # Status
        [Parameter()]
        [ValidateSet('ENABLED','DISABLED')]
        [string]
        $Status,

        # Enable auto registration
        [Parameter()]
        [switch]
        $AutoRegistrationEnabled
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificates update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devicemanagement certificates update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devicemanagement certificates update $c8yargs
        }
        
    }

    End {}
}
