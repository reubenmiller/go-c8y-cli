---
category: Tenants
external help file: PSc8y-help.xml
id: Expand-Tenant
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Tenants/expand-tenant
title: Expand-Tenant
---



## SYNOPSIS
Expand the tenants by id or name

## SYNTAX

```
Expand-Tenant
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
Expand a list of tenants replacing any ids or names with the actual tenant object.

## EXAMPLES

### EXAMPLE 1
```
Expand-Tenant "mytenant"
```

Retrieve the tenant objects by name or id

### EXAMPLE 2
```
Get-Tenant *test* | Expand-Tenant
```

Get all the tenant object (with app in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

## PARAMETERS

### -InputObject
List of ids, names or tenant objects

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
If the given object is already an tenant object, then it is added with no additional lookup

## RELATED LINKS
