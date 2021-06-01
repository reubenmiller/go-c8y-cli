---
category: Applications
external help file: PSc8y-help.xml
id: Expand-Application
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Applications/expand-application
title: Expand-Application
---



## SYNOPSIS
Expand a list of applications replacing any ids or names with the actual application object.

## SYNTAX

```
Expand-Application
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
The list of applications will be expanded to include the full application representation by fetching
the data from Cumulocity.

## EXAMPLES

### EXAMPLE 1
```
Expand-Application "app-name"
```

Retrieve the application objects by name or id

### EXAMPLE 2
```
Get-C8yApplicationCollection *app* | Expand-Application
```

Get all the application object (with app in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

### EXAMPLE 3
```
Expand-Application * -Type MICROSERVICE
```

Expand applications that match a name of "*" and have a type of "MICROSERVICE"

## PARAMETERS

### -InputObject
List of ids, names or application objects

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
If the given object is already an application object, then it is added with no additional lookup

## RELATED LINKS
