---
category: Misc
external help file: PSc8y-help.xml
id: ConvertTo-NestedJson
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/convertto-nestedjson
title: ConvertTo-NestedJson
---



## SYNOPSIS
Convert object to JSON

## SYNTAX

```
ConvertTo-NestedJson
	[-InputObject] <Object>
	[-Depth <Int32>]
	[-Compress]
	[<CommonParameters>]
```

## DESCRIPTION
Convert object to JSON but increase the default depth used by ConvertTo-Json

## EXAMPLES

### EXAMPLE 1
```
@{example = "one"} | ConvertTo-NestedJson
```

Convert object to JSON

## PARAMETERS

### -InputObject
Input object

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

### -Depth
Max depth

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 20
Accept pipeline input: False
Accept wildcard characters: False
```

### -Compress
Compress

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
