# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAssetReference {
<#
.SYNOPSIS
Get managed object child asset reference

.DESCRIPTION
Get managed object child asset reference

.EXAMPLE
PS> Get-ChildAssetReference -Asset $Agent.id -Reference $Ref.id

Get an existing child asset reference


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Asset id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Asset,

        # Asset reference id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Reference,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Reference")) {
            $Parameters["reference"] = PSc8y\Expand-Id $Reference
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        foreach ($item in (PSc8y\Expand-Device $Asset)) {
            if ($item) {
                $Parameters["asset"] = if ($item.id) { $item.id } else { $item }
            }


            Invoke-ClientCommand `
                -Noun "inventoryReferences" `
                -Verb "getChildAsset" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.managedObjectReference+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
