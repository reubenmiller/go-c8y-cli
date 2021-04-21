---
category: Misc.
external help file: PSc8y-help.xml
id: New-TestFile
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc./new-testfile
title: New-TestFile
---



## SYNOPSIS
Create a new temp file with default contents

## SYNTAX

```
New-TestFile
	[[-InputObject] <Object>]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Create a temporary file with some contents which can be used to uploaded it to Cumulocity
via the Binary api.

## EXAMPLES

### EXAMPLE 1
```
New-TestFile
```

Create a temp file with pre-defined content

### EXAMPLE 2
```
"My custom text info" | New-TestFile
```

Create a temp file with customized content.

## PARAMETERS

### -InputObject
Content which should be written to the temporary file

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Example message
Accept pipeline input: True (ByPropertyName, ByValue)
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

### System.IO.FileInfo
## NOTES

## RELATED LINKS
