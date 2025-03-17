---
category: Client Helpers
external help file: PSc8y-help.xml
id: ConvertFrom-ClientOutput
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/convertfrom-clientoutput
title: ConvertFrom-ClientOutput
---



## SYNOPSIS
Convert from the c8y client output into PowerShell objects

## SYNTAX

```
ConvertFrom-ClientOutput
	[-InputObject] <Object[]>
	[-Type <String>]
	[-ItemType <String>]
	[-BoundParameters <Hashtable>]
	[<CommonParameters>]
```

## DESCRIPTION
Convert from the c8y client output into a PowerShell object

## EXAMPLES

### EXAMPLE 1
```
c8y devices list | ConvertFrom-ClientOutput -Type application/json
```

Convert the raw object list into PowerShell objects

## PARAMETERS

### -InputObject
Input Object

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Type
Type

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: Application/json
Accept pipeline input: False
Accept wildcard characters: False
```

### -ItemType
Item type

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: Application/json
Accept pipeline input: False
Accept wildcard characters: False
```

### -BoundParameters
Existing bound parameters from the cmdlet.
Common parameters will be retrieved from it

```yaml
Type: Hashtable
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
