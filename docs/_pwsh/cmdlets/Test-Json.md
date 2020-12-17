---
category: Misc.
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version: https://go.microsoft.com/fwlink/?LinkID=2096609
schema: 2.0.0
title: Test-Json
---

# Test-Json

## SYNOPSIS
Test if the input object is a valid json string

## SYNTAX

```
Test-Json
	[-InputObject] <Object>
	[<CommonParameters>]
```

## DESCRIPTION
Test the given input to check if it is most likely valid json.
The cmdlet uses
a quick json sanity check rather than trying to parse the json to save time.

## EXAMPLES

### EXAMPLE 1
```
Test-Json '{ "name": "tester" }'
```

Returns true if the input data is valid json

## PARAMETERS

### -InputObject
Input data

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByValue)
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

### System.String
## OUTPUTS

### System.Boolean
### System.Object
## NOTES

## RELATED LINKS

[https://go.microsoft.com/fwlink/?LinkID=2096609](https://go.microsoft.com/fwlink/?LinkID=2096609)

