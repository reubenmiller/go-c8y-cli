---
category: Client Helpers
external help file: PSc8y-help.xml
id: Import-ClientBinary
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/import-clientbinary
title: Import-ClientBinary
---



## SYNOPSIS
Import/Download the Cumulocity Binary

## SYNTAX

```
Import-ClientBinary
	[[-Version] <String>]
	[[-Platform] <String>]
	[[-Arch] <String>]
	[-Force]
	[<CommonParameters>]
```

## DESCRIPTION
Get the full path to the Cumulocity Binary which is compatible with the current Operating system

## EXAMPLES

### EXAMPLE 1
```
Import-ClientBinary
```

Download the client binary corresponding to your current platform

## PARAMETERS

### -Version
Version.
Defaults to the module's version

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Platform
OS Platform.
If left blank it will be auto detected

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Arch
CPU Architecture.
If left blank it will be auto detected

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 3
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Force
Force redownloading of the binary

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

### System.String
## NOTES

## RELATED LINKS
