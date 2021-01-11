# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-User {
<#
.SYNOPSIS
Delete user

.DESCRIPTION
Delete a user from the a tenant. This will remove the user completely from the tenant and can not be reversed.
Alternatively a user can be disabled via updating the users properties instead of deleting the user.


.EXAMPLE
PS> Remove-User -Id $User.id

Delete a user


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [string]
        $Id,

        # Tenant
        [Parameter()]
        [object]
        $Tenant,

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "")]
        [string]
        $ProcessingMode,

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
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
        }
        if ($PSBoundParameters.ContainsKey("ProcessingMode")) {
            $Parameters["processingMode"] = $ProcessingMode
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
            # Inherit preference automatic variables
            if (!$WhatIfPreference.IsPresent) {
                $WhatIfPreference = $PSCmdlet.SessionState.PSVariable.get("WhatIfPreference").Value
            }
        
            # Inherit custom parameters
            $Stack = Get-PSCallStack | Select-Object -Skip 1 -First 1
            $InheritVariables = @(@{Source="Force"; Target="Force"}, @{Source="WhatIf"; Target="WhatIfPreference"})
            foreach ($iVariable in $InheritVariables) {
                if (-Not $PSBoundParameters.ContainsKey($iVariable.Source)) {
                    if ($null -ne $Stack -and $Stack.InvocationInfo.BoundParameters.ContainsKey($iVariable.Source)) {
                        Set-Variable -Name $iVariable.Target -Value $Stack.InvocationInfo.BoundParameters[$iVariable.Source] -WhatIf:$false
                    }
                }
            }
        }
    }

    Process {
        foreach ($item in (PSc8y\Expand-Id $Id)) {
            if ($item) {
                $Parameters["id"] = if ($item.id) { $item.id } else { $item }
            }

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "users" `
                -Verb "delete" `
                -Parameters $Parameters `
                -Type "" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
