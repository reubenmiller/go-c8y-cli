---
category: Misc.
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Expand-PaginationObject
---

# Expand-PaginationObject

## SYNOPSIS
Expand a Cumulocity pagination result

## SYNTAX

```
Expand-PaginationObject
	[-InputObject] <Object>
	[[-MaxPage] <Int32>]
	[<CommonParameters>]
```

## DESCRIPTION
{{ Fill in the Description }}

## EXAMPLES

### EXAMPLE 1
```
Invoke-ClientRequest -Uri "/inventory/managedObjects" -QueryParameters @{ pageSize = 2000 } -Raw | ConvertFrom-Json | Expand-PaginationObject
```

Get all managed objects in the platform (rest requests will be done in chunks of 2000)

### EXAMPLE 2
```
$data = Get-MeasurementCollection -Device testDevice -Raw -PageSize 2000 | Expand-PaginationObject
```

Get a measurement collection, then retrieve all the measurements by iterating through the pagination object

## PARAMETERS

### -InputObject
Response from a Cumulocity rest request.
It must have the next property.

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -MaxPage
Maximum number of pages to retrieve.
If Zero or less, then it will retrieve all of the results

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
