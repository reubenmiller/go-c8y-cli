---
category: Misc
external help file: PSc8y-help.xml
id: New-RandomString
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/new-randomstring
title: New-RandomString
---



## SYNOPSIS
Create a random string

## SYNTAX

```
New-RandomString
	[[-Prefix] <String>]
	[[-Postfix] <String>]
	[<CommonParameters>]
```

## DESCRIPTION
Helper utility to quickly create a randomized string which can be used
when adding unique names to devices or another other properties

Note: It should not be used for encryption!

## EXAMPLES

### EXAMPLE 1
```
New-RandomString -Prefix "hello_"
```

Create a random string with the "hello" prefix.
i.e `hello_jta6fzwvo7`

### EXAMPLE 2
```
New-RandomString -Postfix "_device"
```

Create a random string which ends with "_device", i.e.
`1qs7mc2o3t_device`

## PARAMETERS

### -Prefix
Prefix to be added before the random string

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Postfix
Postfix to be added after the random string

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
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
