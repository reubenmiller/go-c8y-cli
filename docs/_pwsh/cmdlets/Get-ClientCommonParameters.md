---
category: Client Helpers
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-ClientCommonParameters
---

# Get-ClientCommonParameters

## SYNOPSIS
Get the common parameters which can be added to a function which extends PSc8y functionality

## SYNTAX

```
Get-ClientCommonParameters
	[-Type] <String[]>
	[-BoundParameters <Hashtable>]
	[<CommonParameters>]
```

## DESCRIPTION
* PageSize

## EXAMPLES

### EXAMPLE 1
```
Function Get-MyObject {
    [cmdletbinding(
        SupportsShouldProcess = $True,
        ConfirmImpact = "None"
    )]
    Param()
```

DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template" -BoundParameters $PSBoundParameters
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

### -BoundParameters
Existing bound parameters from the cmdlet.
Providing it will ensure that the dynamic parameters do not duplicate
existing parameters.

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
