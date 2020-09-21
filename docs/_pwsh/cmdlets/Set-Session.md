---
category: Sessions
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Set-Session
---

# Set-Session

## SYNOPSIS
Set/activate a Cumulocity Session.

## SYNTAX

### None (Default)
```
Set-Session
	[-UseEnvironment]
	[<CommonParameters>]
```

### ByFile
```
Set-Session
	[[-File] <String>]
	[-UseEnvironment]
	[<CommonParameters>]
```

## DESCRIPTION
By default the user will be prompted to select from Cumulocity sessions found in their home folder under .cumulocity

## EXAMPLES

### EXAMPLE 1
```
Set-Session
```

## PARAMETERS

### -File
File containing the Cumulocity session data

```yaml
Type: String
Parameter Sets: ByFile
Aliases: FullName

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -UseEnvironment
Allow loading Cumulocity session setting from environment variables

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

### String
## NOTES

## RELATED LINKS
