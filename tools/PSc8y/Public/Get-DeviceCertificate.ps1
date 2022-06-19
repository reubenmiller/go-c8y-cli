# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceCertificate {
<#
.SYNOPSIS
Get trusted device certificate

.DESCRIPTION
Get a trusted device certificate

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificates_get

.EXAMPLE
PS> Get-DeviceCertificate -Id abcedef0123456789abcedef0123456789

Get trusted device certificate by id/fingerprint


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
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificates get"
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
            | c8y devicemanagement certificates get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devicemanagement certificates get $c8yargs
        }
        
    }

    End {}
}
