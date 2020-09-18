---
category: Microservices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Expand-Microservice
---

# Expand-Microservice

## SYNOPSIS
Expand a list of microservices replacing any ids or names with the actual microservice object.

## SYNTAX

```
Expand-Microservice
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
{{ Fill in the Description }}

## EXAMPLES

### EXAMPLE 1
```
Expand-Microservice "app-name"
```

Retrieve the microservice objects by name or id

### EXAMPLE 2
```
Get-C8yMicroserviceCollection *app* | Expand-Microservice
```

Get all the microservice object (with app in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

## PARAMETERS

### -InputObject
List of ids, names or microservice objects

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

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES
If the given object is already an microservice object, then it is added with no additional lookup

## RELATED LINKS
