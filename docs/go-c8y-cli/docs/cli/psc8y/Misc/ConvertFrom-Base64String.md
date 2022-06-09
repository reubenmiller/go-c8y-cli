---
category: Misc
external help file: PSc8y-help.xml
id: ConvertFrom-Base64String
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/convertfrom-base64string
title: ConvertFrom-Base64String
---



## SYNOPSIS
Convert a base64 encoded string to UTF8

## SYNTAX

```
ConvertFrom-Base64String
	[-InputObject] <String[]>
	[<CommonParameters>]
```

## DESCRIPTION
Convert a base64 encoded string to UTF8

## EXAMPLES

### EXAMPLE 1
```
ConvertFrom-Base64String ZWFzdGVyZWdn
```

Convert the base64 to utf8

### EXAMPLE 2
```
ConvertFrom-Base64String "Authorization: Basic s7sd81kkzyzldjkzkhejhug3kh"
```

Convert the base64 to utf8

## PARAMETERS

### -InputObject
Base64 encoded string

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
If the the string has spaces in it, then only the last part of the string (with no spaces in it) will be used.
This makes it easier when trying decode the basic auth string

## RELATED LINKS
