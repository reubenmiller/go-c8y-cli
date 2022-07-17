---
category: Misc
external help file: PSc8y-help.xml
id: Expand-Id
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/expand-id
title: Expand-Id
---



## SYNOPSIS
Expand a list of ids.

## SYNTAX

```
Expand-Id
	[-InputObject] <Object[]>
	[-Strict]
	[<CommonParameters>]
```

## DESCRIPTION
Expand the list of objects and only return the ids instead of the full objects.

## EXAMPLES

### EXAMPLE 1
```
Expand-Id 12345
```

Normalize a list of ids

### EXAMPLE 2
```
"12345", "56789" | Expand-Id
```

Normalize a list of ids

## PARAMETERS

### -InputObject
List of ids

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByValue)
Accept wildcard characters: False
```

### -Strict
Exclude all non-id like values

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
