# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-User {
<#
.SYNOPSIS
Update user

.DESCRIPTION
Update properties, reset password or enable/disable for a user in a tenant

.EXAMPLE
PS> Update-User -Id $User.id -FirstName "Simon"

Update a user


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
        [object[]]
        $Id,

        # User first name
        [Parameter()]
        [string]
        $FirstName,

        # User last name
        [Parameter()]
        [string]
        $LastName,

        # User phone number. Format: '+[country code][number]', has to be a valid MSISDN
        [Parameter()]
        [string]
        $Phone,

        # User email address
        [Parameter()]
        [string]
        $Email,

        # User activation status (true/false)
        [Parameter()]
        [switch]
        $Enabled,

        # User password. Min: 6, max: 32 characters. Only Latin1 chars allowed
        [Parameter()]
        [string]
        $Password,

        # User activation status (true/false)
        [Parameter()]
        [ValidateSet('true','false')]
        [switch]
        $SendPasswordResetEmail,

        # Custom properties to be added to the user
        [Parameter()]
        [object]
        $CustomProperties,

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

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

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
        if ($PSBoundParameters.ContainsKey("FirstName")) {
            $Parameters["firstName"] = $FirstName
        }
        if ($PSBoundParameters.ContainsKey("LastName")) {
            $Parameters["lastName"] = $LastName
        }
        if ($PSBoundParameters.ContainsKey("Phone")) {
            $Parameters["phone"] = $Phone
        }
        if ($PSBoundParameters.ContainsKey("Email")) {
            $Parameters["email"] = $Email
        }
        if ($PSBoundParameters.ContainsKey("Enabled")) {
            $Parameters["enabled"] = $Enabled
        }
        if ($PSBoundParameters.ContainsKey("Password")) {
            $Parameters["password"] = $Password
        }
        if ($PSBoundParameters.ContainsKey("SendPasswordResetEmail")) {
            $Parameters["sendPasswordResetEmail"] = $SendPasswordResetEmail
        }
        if ($PSBoundParameters.ContainsKey("CustomProperties")) {
            $Parameters["customProperties"] = ConvertTo-JsonArgument $CustomProperties
        }
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
        }
        if ($PSBoundParameters.ContainsKey("ProcessingMode")) {
            $Parameters["processingMode"] = $ProcessingMode
        }
        if ($PSBoundParameters.ContainsKey("Template") -and $Template) {
            $Parameters["template"] = $Template
        }
        if ($PSBoundParameters.ContainsKey("TemplateVars") -and $TemplateVars) {
            $Parameters["templateVars"] = $TemplateVars
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
        foreach ($item in (PSc8y\Expand-User $Id)) {
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
                -Verb "update" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.user+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
