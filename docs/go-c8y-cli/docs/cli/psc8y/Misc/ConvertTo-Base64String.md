---
category: Misc
external help file: PSc8y-help.xml
id: ConvertTo-Base64String
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/convertto-base64string
title: ConvertTo-Base64String
---



## SYNOPSIS
Convert a UTF8 string to a base64 encoded string

## SYNTAX

```
ConvertTo-Base64String
	[-InputObject] <String[]>
	[<CommonParameters>]
```

## DESCRIPTION
Convert a UTF8 string to a base64 encoded string

## EXAMPLES

### EXAMPLE 1
```
ConvertTo-Base64String tenant/username:password
```

Encode a string to base64 encoded string

## PARAMETERS

### -InputObject
UTF8 encoded string

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
