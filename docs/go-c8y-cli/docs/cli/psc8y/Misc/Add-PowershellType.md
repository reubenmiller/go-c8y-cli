---
category: Misc
external help file: PSc8y-help.xml
id: Add-PowershellType
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/add-powershelltype
title: Add-PowershellType
---



## SYNOPSIS
Add a powershell type name to a powershell object

## SYNTAX

```
Add-PowershellType
	-InputObject <Object[]>
	[-Type] <String>
	[<CommonParameters>]
```

## DESCRIPTION
This allows a custom type name to be given to powershell objects, so that the view formatting can be used (i.e.
.ps1xml)

## EXAMPLES

### EXAMPLE 1
```
$data | Add-PowershellType -Type "customType1"
```

Add a type `customType1` to the input object

## PARAMETERS

### -InputObject
Object to add the type name to

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: Named
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Type
Type name to assign to the input objects

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: 2
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

### Object[]
## OUTPUTS

### Object[]
## NOTES

## RELATED LINKS
