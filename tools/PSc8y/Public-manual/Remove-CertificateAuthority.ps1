Function Remove {
<#
.SYNOPSIS
(PREVIEW FEATURE) Delete tenant certificate authority

.DESCRIPTION
Delete the tenant's Certificate Authority used for device onboarding


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificate-authority_delete

.EXAMPLE
PS> Remove-CertificateAuthority

Delete the tenant's certificate authority


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificate-authority delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicemanagement certificate-authority delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicemanagement certificate-authority delete $c8yargs
        }
    }

    End {}
}
