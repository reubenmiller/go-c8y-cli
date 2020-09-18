---
category: Client Helpers
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Set-ClientConsoleSetting
---

# Set-ClientConsoleSetting

## SYNOPSIS
Set console settings to be used by the cli tool

## SYNTAX

```
Set-ClientConsoleSetting
	[-HideSensitive]
	[-ShowSensitive]
	[<CommonParameters>]
```

## DESCRIPTION
Sensitive information:
When using -HideSensitive, the following information will be obfuscated when shown on the console
(tenant, username, password, base64 credentials)

## EXAMPLES

### EXAMPLE 1
```
Set-ClientConsoleSetting -HideSensitive
```

Hide any sensitive session information on the console.
Settings like (tenant, username, password, base64 credentials)

## PARAMETERS

### -HideSensitive
Hide all sensitive session information (tenant, username, password, base64 encoded passwords etc.)

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

### -ShowSensitive
Show sensitive information (excepts clear-text passwords)

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

## NOTES

## RELATED LINKS
