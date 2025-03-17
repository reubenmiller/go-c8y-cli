Function Invoke-NativeCumulocityRequest {
<#
.SYNOPSIS
Invoke a native Cumulocity Request using only PowerShell

.DESCRIPTION
Invoke a native Cumulocity Request using only PowerShell

.EXAMPLE
Invoke-NativeCumulocityRequest -Uri inventory/managedObjects

Send a REST request to inventory/managedObjects
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Uri
        [Alias("Url")]
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Uri,

        # Method
        [string] $Method,

        # Body
        [object] $Body,

        # Headers
        [object] $Headers
    )

    Begin {
        $FullUri = $Uri
        if (!$FullUri.StartsWith("http")) {
            $FullUri = @($env:C8Y_URL, $Uri.TrimStart("/")) -join "/"
        }

        $options = @{
            Uri = $FullUri
        }

        if ($PSBoundParameters.ContainsKey("Method")) {
            $options.Method = $Method
        }

        $AllHeaders = @{}

        if ($PSBoundParameters.ContainsKey("Headers")) {
            $AllHeaders = @{} + $Headers
        }

        if ($Env:C8Y_TOKEN) {
            $AllHeaders.Authorization = "Bearer " + $env:C8Y_TOKEN
        } else {
            $AllHeaders.Authorization = "Basic " + (ConvertTo-Base64String ("{0}/{1}:{2}" -f $env:C8Y_TENANT, $env:C8Y_USERNAME, $env:C8Y_PASSWORD))
        }
    }

    Process {

        if ($PSBoundParameters.ContainsKey("Body")) {
            $options.Body = $Body
        }

        $options.Headers = $AllHeaders
        Invoke-RestMethod @options
    }
}