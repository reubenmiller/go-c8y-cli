Function New-CertificateAuthority {
<#
.SYNOPSIS
(PREVIEW FEATURE) Create tenant certificate authority

.DESCRIPTION
Create a key pair and self-sign a certificate with as the Common Name (CN).
Store the private key in an encrypted tenant option. Store the certificate in
the trusted certificate repository with auto-registration unchecked by default.
The devices can be registered automatically only when device administrator checks
this option ON. If the CA certificate is removed from the trusted certificate list,
corresponding public and private key removed automatically from the database collection.
If a CA is already present, return a message indicating the CA is already present.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificate-authority_create

.EXAMPLE
PS> New-CertificateAuthority

Create new certificate authority


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificate-authority create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicemanagement certificate-authority create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicemanagement certificate-authority create $c8yargs
        }
    }

    End {}
}
