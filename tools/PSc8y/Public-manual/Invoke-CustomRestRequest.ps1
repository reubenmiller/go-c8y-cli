Function Invoke-CustomRestRequest {
    [cmdletbinding()]
    Param(
        [string] $Uri,

        [string] $Method,

        [hashtable] $QueryParameters
    )


    $pair = "{0}:{1}" -f $env:C8Y_USER, $env:C8Y_PASSWORD

    $Headers = @{
        Authorization = "Basic {0}" -f [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($pair))
    }

    Invoke-WebRequest -Uri:$Uri -Method:$Method -Headers:$Headers -UserAgent "pscumulocity.v2"

}
