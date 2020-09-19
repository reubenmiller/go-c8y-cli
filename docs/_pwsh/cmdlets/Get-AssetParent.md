---
category: Assets
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-AssetParent
---

# Get-AssetParent

## SYNOPSIS
Get asset parent references for a asset

## SYNTAX

### ByLevel (Default)
```
Get-AssetParent
	[[-Asset] <Object[]>]
	[[-Level] <Int32>]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

### Root
```
Get-AssetParent
	[[-Asset] <Object[]>]
	[-RootParent]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

### All
```
Get-AssetParent
	[[-Asset] <Object[]>]
	[-All]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Get the parent of an existing assert by using the references.
The cmdlet supports returning
various forms of the parent references, i.e.
immediate parent, parent or the parent, or the
full parental references.

## EXAMPLES

### EXAMPLE 1
```
Get-AssetParent asset0*
```

Get the direct (immediate) parent of the given asset

### EXAMPLE 2
```
Get-AssetParent -All
```

Return an array of parent assets where the first element in the array is the root asset, and the last is the direct parent of the given asset.

### EXAMPLE 3
```
Get-AssetParent -RootParent
```

Returns the root parent.
In most cases this will be the agent

## PARAMETERS

### -Asset
Asset id, name or object.
Wildcards accepted

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Level
Level to navigate backward from the given asset to its parent/s
1 = direct parent
2 = parent of its parent
If the Level is too large, then the root parent will be returned

```yaml
Type: Int32
Parameter Sets: ByLevel
Aliases:

Required: False
Position: 2
Default value: 1
Accept pipeline input: False
Accept wildcard characters: False
```

### -RootParent
Return the top level / root parent

```yaml
Type: SwitchParameter
Parameter Sets: Root
Aliases:

Required: False
Position: Named
Default value: False
Accept pipeline input: False
Accept wildcard characters: False
```

### -All
Return a list of all parent assets

```yaml
Type: SwitchParameter
Parameter Sets: All
Aliases:

Required: False
Position: Named
Default value: False
Accept pipeline input: False
Accept wildcard characters: False
```

### -WhatIf
Shows what would happen if the cmdlet runs.
The cmdlet is not run.

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: wi

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Confirm
Prompts you for confirmation before running the cmdlet.

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: cf

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
