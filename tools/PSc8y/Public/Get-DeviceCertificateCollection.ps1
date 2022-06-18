# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceCertificateCollection {
<#
.SYNOPSIS
List device certificates

.DESCRIPTION
List the trusted device certificates


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicemanagement_certificates_list

.EXAMPLE
PS> Get-DeviceCertificateCollection

Get list of trusted device certificates


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicemanagement certificates list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Tenant `
            | Group-ClientRequests `
            | c8y devicemanagement certificates list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y devicemanagement certificates list $c8yargs
        }
        
    }

    End {}
}
