---
category: Devices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-DeviceBootstrapCredential
---

# Get-DeviceBootstrapCredential

## SYNOPSIS
Get the device bootstrap credential as a PowerShell credential object (for use in Rest requests)

## SYNTAX

```
Get-DeviceBootstrapCredential
	[<CommonParameters>]
```

## DESCRIPTION
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
$Credential.GetPlainText()  # =\> returns credentials in format "{username}/{password}"
$Credential.GetBasicAuth()  # =\> returns credentials in format "Basic {base64 encoded username/password}"
```

The credentials can be obtained by contacting support.
For security reasons, do not use your tenant credentials.

## EXAMPLES

### EXAMPLE 1
```
New-DeviceBootstrapCredential
```

Get a credential object containing the devicebootstrap credentials

### EXAMPLE 2
```
$Cred = New-DeviceBootstrapCredential; $Cred.GetBasicAuth()
```

Get device bootstrap credentials in the format of basic auth (for use in the 'Authorization' header)

## PARAMETERS

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

### System.Management.Automation.PSCredential
## NOTES

## RELATED LINKS
