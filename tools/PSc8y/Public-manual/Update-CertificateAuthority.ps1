Function Update-CertificateAuthority {
<#
.SYNOPSIS
Update tenant's certificate authority

.DESCRIPTION
Update tenant's certificate authority

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificate-authority_update

.EXAMPLE
PS> Update-CertificateAuthority -Status DISABLED

Disable the tenant's certificate authority

.EXAMPLE
PS> Update-CertificateAuthority -AutoRegistrationEnabled:$false

Disable auto registration on the tenant's certificate authority to prevent new devices from being registered

#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificate-authority update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicemanagement certificate-authority update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicemanagement certificate-authority update $c8yargs
        }
        
    }

    End {}
}
