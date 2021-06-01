---
category: Client Helpers
external help file: PSc8y-help.xml
id: Get-ClientCommonParameters
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/get-clientcommonparameters
title: Get-ClientCommonParameters
---



## SYNOPSIS
Get the common parameters which can be added to a function which extends PSc8y functionality

## SYNTAX

```
Get-ClientCommonParameters
	[-Type] <String[]>
	[-SkipConfirm]
	[<CommonParameters>]
```

## DESCRIPTION
* PageSize

## EXAMPLES

### EXAMPLE 1
```
Function Get-MyObject {
    [cmdletbinding()]
    Param()
```

DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        Find-ManagedObjects @PSBoundParameters
    }
}
Inherit common parameters to a custom function.
This will add parameters such as "PageSize", "TotalPages", "Template" to your function

## PARAMETERS

### -Type
Parameter types to include

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -SkipConfirm
Ignore confirm parameter (i.e.
when using inbuilt powershell -Confirm)

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
