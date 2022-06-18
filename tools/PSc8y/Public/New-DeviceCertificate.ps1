# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DeviceCertificate {
<#
.SYNOPSIS
Upload trusted device certificate

.DESCRIPTION
Upload a trusted device certificate which will enable communication to Cumulocity using the certificate (or a cert which is trusted by the certificate)


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificates_create

.EXAMPLE
PS> New-DeviceCertificate -Name "MyCert" -File "./cert.pem"

Upload a trusted device certificate


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id (required)
        [Parameter(Mandatory = $true)]
        [object]
        $Tenant,

        # Certificate name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Status (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('ENABLED','DISABLED')]
        [string]
        $Status,

        # Status
        [Parameter()]
        [string]
        $File,

        # Enable auto registration
        [Parameter()]
        [switch]
        $AutoRegistrationEnabled
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificates create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y devicemanagement certificates create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y devicemanagement certificates create $c8yargs
        }
        
    }

    End {}
}
