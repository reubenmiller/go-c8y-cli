---
category: Sessions
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-Session
---

# New-Session

## SYNOPSIS
Create a new Cumulocity Session

## SYNTAX

```
New-Session
	-Name <String>
	-Host <String>
	-Tenant <String>
	[-Credential <PSCredential>]
	[-Description <String>]
	[-NoTenantPrefix]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new Cumulocity session which can be used by the cmdlets.
The new session will be automatically activated.

## EXAMPLES

### EXAMPLE 1
```
New-Session -Name "develop" -Host "https://my-tenant-name.eu-latest.cumulocity.com" -Tenant "t12345"
```

Create a new Cumulocity session

## PARAMETERS

### -Name
Name of the Cumulocity session

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Host
Host url, i.e.
https://my-tenant-name.eu-latest.cumulocity.com

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Tenant
Tenant id, i.e.
t12345

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Credential
Credential

```yaml
Type: PSCredential
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: [System.Management.Automation.PSCredential]::Empty
Accept pipeline input: False
Accept wildcard characters: False
```

### -Description
Description

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoTenantPrefix
Don't use tenant name as a prefix to the user name when using Basic Authentication

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: False
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

### None
## NOTES

## RELATED LINKS
