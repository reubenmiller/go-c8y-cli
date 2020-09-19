Function Get-DeviceBootstrapCredential {
<#
.SYNOPSIS
Get the device bootstrap credential as a PowerShell credential object (for use in Rest requests)

.DESCRIPTION
The PSCredentials object also has two additional methods to make the usage of the credentials easier in

The device bootstrap credentials should be already set in the following environment variables

```powershell
$env:C8Y_DEVICEBOOTSTRAP_TENANT
$env:C8Y_DEVICEBOOTSTRAP_USERNAME
$env:C8Y_DEVICEBOOTSTRAP_PASSWORD
```

Then the credentials can be retrieved using

```powershell
$Credential = Get-DeviceBootstrapCredential
$Credential.GetPlainText()  # => returns credentials in format "{username}/{password}"
$Credential.GetBasicAuth()  # => returns credentials in format "Basic {base64 encoded username/password}"
```

The credentials can be obtained by contacting support. For security reasons, do not use your tenant credentials.

.OUTPUTS
System.Management.Automation.PSCredential

.EXAMPLE
New-DeviceBootstrapCredential

Get a credential object containing the devicebootstrap credentials

.EXAMPLE
$Cred = New-DeviceBootstrapCredential; $Cred.GetBasicAuth()

Get device bootstrap credentials in the format of basic auth (for use in the 'Authorization' header)
#>
    [cmdletbinding()]
    Param()

    $ErrorMessages = New-Object System.Collections.ArrayList
    if (!$env:C8Y_DEVICEBOOTSTRAP_TENANT) {
        $null = $ErrorMessages.Add("Missing env variable: C8Y_DEVICEBOOTSTRAP_TENANT")
    }
    if (!$env:C8Y_DEVICEBOOTSTRAP_USERNAME) {
        $null = $ErrorMessages.Add("Missing env variable: C8Y_DEVICEBOOTSTRAP_USERNAME")
    }
    if (!$env:C8Y_DEVICEBOOTSTRAP_PASSWORD) {
        $null = $ErrorMessages.Add("Missing env variable: C8Y_DEVICEBOOTSTRAP_PASSWORD")
    }

    if ($ErrorMessages.Count -ne 0) {
        Write-Warning ("The following environment variables are missing:`n    {0}" -f ($ErrorMessages -join "`n    "))
        return
    }

    # Get credentials from environment variables
    $Tenant = $env:C8Y_DEVICEBOOTSTRAP_TENANT
    $Username = $env:C8Y_DEVICEBOOTSTRAP_USERNAME
    $Password = $env:C8Y_DEVICEBOOTSTRAP_PASSWORD | ConvertTo-SecureString -asPlainText -Force

    $Credential = New-Object System.Management.Automation.PSCredential("$Tenant/$Username", $password)

    # Add helper to return clear text username/password
    $Credential | Add-Member -MemberType ScriptMethod -Name "GetPlainText" -Value {
        "{0}:{1}" -f $this.GetNetworkCredential().UserName, $this.GetNetworkCredential().Password
    }

    # Add helper to return basic auth header info
    $Credential | Add-Member -MemberType ScriptMethod -Name "GetBasicAuth" -Value {
        "Basic " + [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes(("{0}:{1}" -f @(
            $this.GetNetworkCredential().UserName,
            $this.GetNetworkCredential().Password
        ))))
    }

    $Credential
}
