---
category: User Groups
external help file: PSc8y-help.xml
id: Group-ClientRequests
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/User Groups/group-clientrequests
title: Group-ClientRequests
---



## SYNOPSIS
Groups the input into array of a given maximum size.

## SYNTAX

```
Group-ClientRequests
	[-InputObject] <Object[]>
	[-Size <Int32>]
	[-AsPSObject]
	[<CommonParameters>]
```

## DESCRIPTION
Groups the input into array of a given maximum size.
It will pass the piped input as array rather than individual items
This cmdlet is mainly used internally by the module.

## EXAMPLES

### EXAMPLE 1
```
$Id | Group-ClientRequests | c8y devices delete $c8yargs
```

Group and normalize the input objects to be compatible with piping to the native c8y binary

## PARAMETERS

### -InputObject
Input objects to be piped to native c8y binary

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Size
Grouping size

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 2000
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsPSObject
Output objects as PSObjects rather than json text

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
