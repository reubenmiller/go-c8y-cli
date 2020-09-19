Function New-RandomPassword {
<#
.SYNOPSIS
Create pseudo random password

.DESCRIPTION
Create a random password which can be used for one-time passwords if the
the password reset functionilty in Cumulocity is not available.

.EXAMPLE
New-RandomPassword

C&3SX6Kn

Generate one password with a length between 8 and 12 chars.

.EXAMPLE
New-RandomPassword -MinPasswordLength 8 -MaxPasswordLength 12 -Count 4

7d&5cnaB
!Bh776T"Fw
9"C"RxKcY
%mtM7#9LQ9h

Generate four passwords, each with a length of between 8 and 12 chars.

.EXAMPLE
New-RandomPassword -InputStrings abc, ABC, 123 -PasswordLength 4

3ABa

Generate a password with a length of 4 containing atleast one char from each InputString

.EXAMPLE
New-RandomPassword -InputStrings abc, ABC, 123 -PasswordLength 4 -FirstChar abcdefghijkmnpqrstuvwxyzABCEFGHJKLMNPQRSTUVWXYZ
3ABa

Generates a password with a length of 4 containing atleast one char from each InputString that will start with a letter from 
the string specified with the parameter FirstChar

.OUTPUTS
[String]


.FUNCTIONALITY
Generates random passwords

.LINK
http://blog.simonw.se/powershell-generating-random-password-for-active-directory/

    #>
    [CmdletBinding(
        DefaultParameterSetName = 'FixedLength',
        ConfirmImpact = 'None')]
    [OutputType([String])]
    Param
    (
        # Specifies minimum password length
        [Parameter(Mandatory = $false,
            ParameterSetName = 'RandomLength')]
        [ValidateScript( { $_ -gt 0 })]
        [Alias('Min')] 
        [int]$MinPasswordLength = 12,
        
        # Specifies maximum password length
        [Parameter(Mandatory = $false,
            ParameterSetName = 'RandomLength')]
        [ValidateScript( {
                if ($_ -ge $MinPasswordLength) { $true }
                else { Throw 'Max value cannot be lesser than min value.' } })]
        [Alias('Max')]
        [int]$MaxPasswordLength = 20,

        # Specifies a fixed password length
        [Parameter(Mandatory = $false,
            ParameterSetName = 'FixedLength')]
        [ValidateRange(1, 2147483647)]
        [int]$PasswordLength = 12,
        
        # Specifies an array of strings containing charactergroups from which the password will be generated.
        # At least one char from each group (string) will be used.
        [String[]]$InputStrings = @('abcdefghijkmnpqrstuvwxyz', 'ABCEFGHJKLMNPQRSTUVWXYZ', '123456789', '!#%()[]*+-_;,.'),

        # Specifies a string containing a character group from which the first character in the password will be generated.
        # Useful for systems which requires first char in password to be alphabetic.
        [String] $FirstChar,
        
        # Specifies number of passwords to generate.
        [ValidateRange(1, 2147483647)]
        [int]$Count = 1
    )
    Begin {
        Function Get-Seed {
            # Generate a seed for randomization
            $RandomBytes = New-Object -TypeName 'System.Byte[]' 4
            $Random = New-Object -TypeName 'System.Security.Cryptography.RNGCryptoServiceProvider'
            $Random.GetBytes($RandomBytes)
            [BitConverter]::ToUInt32($RandomBytes, 0)
        }
    }
    Process {
        For ($iteration = 1; $iteration -le $Count; $iteration++) {
            $Password = @{ }
            # Create char arrays containing groups of possible chars
            [char[][]]$CharGroups = $InputStrings

            # Create char array containing all chars
            $AllChars = $CharGroups | ForEach-Object { [Char[]]$_ }

            # Set password length
            if ($PSCmdlet.ParameterSetName -eq 'RandomLength') {
                if ($MinPasswordLength -eq $MaxPasswordLength) {
                    # If password length is set, use set length
                    $PasswordLength = $MinPasswordLength
                }
                else {
                    # Otherwise randomize password length
                    $PasswordLength = ((Get-Seed) % ($MaxPasswordLength + 1 - $MinPasswordLength)) + $MinPasswordLength
                }
            }

            # If FirstChar is defined, randomize first char in password from that string.
            if ($PSBoundParameters.ContainsKey('FirstChar')) {
                $Password.Add(0, $FirstChar[((Get-Seed) % $FirstChar.Length)])
            }
            # Randomize one char from each group
            Foreach ($Group in $CharGroups) {
                if ($Password.Count -lt $PasswordLength) {
                    $Index = Get-Seed
                    While ($Password.ContainsKey($Index)) {
                        $Index = Get-Seed                        
                    }
                    $Password.Add($Index, $Group[((Get-Seed) % $Group.Count)])
                }
            }

            # Fill out with chars from $AllChars
            for ($i = $Password.Count; $i -lt $PasswordLength; $i++) {
                $Index = Get-Seed
                While ($Password.ContainsKey($Index)) {
                    $Index = Get-Seed                        
                }
                $Password.Add($Index, $AllChars[((Get-Seed) % $AllChars.Count)])
            }
            Write-Output -InputObject $( -join ($Password.GetEnumerator() | Sort-Object -Property Name | Select-Object -ExpandProperty Value))
        }
    }
}