---
category: Misc
external help file: PSc8y-help.xml
id: New-RandomPassword
Module Name: PSc8y
online version: http://blog.simonw.se/powershell-generating-random-password-for-active-directory/
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/new-randompassword
title: New-RandomPassword
---



## SYNOPSIS
Create pseudo random password

## SYNTAX

### FixedLength (Default)
```
New-RandomPassword
	[-PasswordLength <Int32>]
	[-InputStrings <String[]>]
	[-FirstChar <String>]
	[-Count <Int32>]
	[<CommonParameters>]
```

### RandomLength
```
New-RandomPassword
	[-MinPasswordLength <Int32>]
	[-MaxPasswordLength <Int32>]
	[-InputStrings <String[]>]
	[-FirstChar <String>]
	[-Count <Int32>]
	[<CommonParameters>]
```

## DESCRIPTION
Create a random password which can be used for one-time passwords if the
the password reset functionilty in Cumulocity is not available.

## EXAMPLES

### EXAMPLE 1
```
New-RandomPassword
```

C&3SX6Kn

Generate one password with a length between 8 and 12 chars.

### EXAMPLE 2
```
New-RandomPassword -MinPasswordLength 8 -MaxPasswordLength 12 -Count 4
```

7d&5cnaB
!Bh776T"Fw
9"C"RxKcY
%mtM7#9LQ9h

Generate four passwords, each with a length of between 8 and 12 chars.

### EXAMPLE 3
```
New-RandomPassword -InputStrings abc, ABC, 123 -PasswordLength 4
```

3ABa

Generate a password with a length of 4 containing atleast one char from each InputString

### EXAMPLE 4
```
New-RandomPassword -InputStrings abc, ABC, 123 -PasswordLength 4 -FirstChar abcdefghijkmnpqrstuvwxyzABCEFGHJKLMNPQRSTUVWXYZ
3ABa
```

Generates a password with a length of 4 containing atleast one char from each InputString that will start with a letter from 
the string specified with the parameter FirstChar

## PARAMETERS

### -MinPasswordLength
Specifies minimum password length

```yaml
Type: Int32
Parameter Sets: RandomLength
Aliases: Min

Required: False
Position: Named
Default value: 12
Accept pipeline input: False
Accept wildcard characters: False
```

### -MaxPasswordLength
Specifies maximum password length

```yaml
Type: Int32
Parameter Sets: RandomLength
Aliases: Max

Required: False
Position: Named
Default value: 20
Accept pipeline input: False
Accept wildcard characters: False
```

### -PasswordLength
Specifies a fixed password length

```yaml
Type: Int32
Parameter Sets: FixedLength
Aliases:

Required: False
Position: Named
Default value: 12
Accept pipeline input: False
Accept wildcard characters: False
```

### -InputStrings
Specifies an array of strings containing charactergroups from which the password will be generated.
At least one char from each group (string) will be used.

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: @('abcdefghijkmnpqrstuvwxyz', 'ABCEFGHJKLMNPQRSTUVWXYZ', '123456789', '!#%()[]*+-_;,.')
Accept pipeline input: False
Accept wildcard characters: False
```

### -FirstChar
Specifies a string containing a character group from which the first character in the password will be generated.
Useful for systems which requires first char in password to be alphabetic.

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

### -Count
Specifies number of passwords to generate.

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 1
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

### [String]
## NOTES

## RELATED LINKS

[http://blog.simonw.se/powershell-generating-random-password-for-active-directory/](http://blog.simonw.se/powershell-generating-random-password-for-active-directory/)

