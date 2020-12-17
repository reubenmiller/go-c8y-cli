---
category: Client Helpers
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Test-ClientPassphrase
---

# Test-ClientPassphrase

## SYNOPSIS
Test the client passphrase which is used to encrypt sensitive Cumulocity session information.

## SYNTAX

```
Test-ClientPassphrase
	[<CommonParameters>]
```

## DESCRIPTION
The passphrase is used to encrypt sensitive information such as passwords and authorization cookies.

The passphrase is saved in an environment variable where it used for all c8y commands to decrypt
the sensitive information.

## EXAMPLES

### EXAMPLE 1
```
Test-ClientPassphrase
```

Set the passphrase if it is not already set

## PARAMETERS

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
