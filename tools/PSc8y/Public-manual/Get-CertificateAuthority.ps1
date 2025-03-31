Function Get-CertificateAuthority {
<#
.SYNOPSIS
(PREVIEW FEATURE) Get tenant certificate authority

.DESCRIPTION
Get the tenant's Certificate Authority used for device onboarding


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificate-authority_get

.EXAMPLE
PS> Get-CertificateAuthority

Get the tenant's certificate authority


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificate-authority get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicemanagement certificate-authority get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicemanagement certificate-authority get $c8yargs
        }
    }

    End {}
}
