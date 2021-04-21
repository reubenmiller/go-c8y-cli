---
category: Client Helpers
external help file: PSc8y-help.xml
id: Install-ClientBinary
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/install-clientbinary
title: Install-ClientBinary
---



## SYNOPSIS
Install the Cumulocity cli binary (c8y)

## SYNTAX

```
Install-ClientBinary
	[[-InstallPath] <String>]
	[<CommonParameters>]
```

## DESCRIPTION
Install the Cumulocity cli binary (c8y) so it is accessible from everywhere in consoles (assuming /usr/local/bin is in the $PATH variable)

## EXAMPLES

### EXAMPLE 1
```
Install-ClientBinary
```

On Linux/MacOS, this installs the cumulocity binary to /usr/local/bin
On Windows this will throw a warning

### EXAMPLE 2
```
Install-ClientBinary -InstallPath /usr/bin
```

Install the Cumulocity binary to /usr/bin

## PARAMETERS

### -InstallPath
Cumulocity installation path where the c8y binaries will be installed.
Defaults to $env:C8Y_INSTALL_PATH

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: $env:C8Y_INSTALL_PATH
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
