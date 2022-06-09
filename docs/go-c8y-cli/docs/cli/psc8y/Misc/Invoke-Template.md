---
category: Misc.
external help file: PSc8y-help.xml
id: Invoke-Template
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc./invoke-template
title: Invoke-Template
---



## SYNOPSIS
Execute a jsonnet data template

## SYNTAX

```
Invoke-Template
	[-Template] <String>
	[-TemplateVars <String>]
	[-Data <Object[]>]
	[-Compact]
	[<CommonParameters>]
```

## DESCRIPTION
Execute a jsonnet data template and show the output of the template.
Useful when developing new templates

## EXAMPLES

### EXAMPLE 1
```
Invoke-Template -Template ./template.jsonnet
```

Execute a jsonnet template

### EXAMPLE 2
```
Invoke-Template -Template ./template.jsonnet -TemplateVars "name=input"
```

Execute a jsonnet template

### EXAMPLE 3
```
Invoke-Template -Template ./template.jsonnet -TemplateVars "name=input,type=mytype"
```

Execute a jsonnet template which has multiple template variables (using a comma separated string)

## PARAMETERS

### -Template
Template (jsonnet) file to use to create the request body.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -TemplateVars
Variables to be used when evaluating the Template.
Accepts a file path, json or json shorthand, i.e.
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

### -Data
Template input data

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Compact
Output compressed/minified json

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: Compress

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

### String
## NOTES

## RELATED LINKS

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/template_execute](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/template_execute)

