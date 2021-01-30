Function Invoke-NativeCumulocityRequest {
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Uri,

        [string] $Method,

        [object] $Body,

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

        $AllCookies = get-item "Env:\C8Y_CREDENTIAL_COOKIES*"

        if ($AllCookies.Count -gt 0) {
            foreach ($iCookie in $AllCookies) {
                $parts = $iCookie.Value -Split "=", 2 | Where-Object { $_ }
                if ($parts.Count -eq 2) {
                    $Allheaders[$parts[0].ToUpper()] = $parts[1]
                }
            }
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