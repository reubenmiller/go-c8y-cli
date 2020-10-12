---
category: Users
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestUser
---

# New-TestUser

## SYNOPSIS
Create a new test user

## SYNTAX

```
New-TestUser
	[[-Name] <String>]
	[-Template <String>]
	[-TemplateVars <String>]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Create a user with a randomized username

## EXAMPLES

### EXAMPLE 1
```
New-TestUser
```

Create a new test user

### EXAMPLE 2
```
New-TestUser -Name "myExistingDevice"
```

Create a new test user with a custom username prefix

## PARAMETERS

### -Name
Name of the username.
A random postfix will be added to it to make it unique

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Testuser
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Template
Template (jsonnet) file to use to create the request body.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -TemplateVars
Variables to be used when evaluating the Template.
Accepts json or json shorthand, i.e.
"name=peter"

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Force
Don't prompt for confirmation

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
