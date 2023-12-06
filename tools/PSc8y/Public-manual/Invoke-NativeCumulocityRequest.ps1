Function Invoke-NativeCumulocityRequest {
<#
.SYNOPSIS
Invoke a Cumulocity IoT request using PowerShell

.DESCRIPTION
Invoke a Cumulocity IoT request using PowerShell

#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Uri to send request to
        [Alias("Url")]
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Uri,

        # HTTP Method, e.g. POST, PUT, GET
        [string] $Method,

        # Request body
        [object] $Body,

        # HTTP Headers to include in the request
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

        $Allheaders = @{}

        if ($PSBoundParameters.ContainsKey("Headers")) {
            $Allheaders = @{} + $Headers
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

        $options.Headers = $Allheaders
        Invoke-RestMethod @options
    }
}