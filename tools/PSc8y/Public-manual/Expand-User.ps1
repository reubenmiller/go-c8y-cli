Function Expand-User {
<#
.SYNOPSIS
Expand a list of users replacing any ids or names with the actual user object.

.NOTES
If the given object is already an user object, then it is added with no additional lookup

.PARAMETER InputObject
List of ids, names or user objects

.EXAMPLE
Expand-User "myuser"

Retrieve the user objects by name or id

.EXAMPLE
Get-UserCollection *test* | Expand-User

Get all the user object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.


#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject
    )

    Process {
        [array] $AllUsers = foreach ($iUser in $InputObject)
        {
            if (($iUser -is [string]))
            {
                if ($iUser -match "\*") {
                    # Remove any wildcard characters, as they are not supported by c8y
                    $iUserNormalized = $iUser -replace "\*", ""

                    PSc8y\Get-UserCollection -Username $iUserNormalized -WhatIf:$false -PageSize 100 |
                        Where-Object { $_.id -like $iUser -or $_.userName -like $iUser }
                } else {
                    $iUser
                }
            }
            else
            {
                $iUser
            }
        }

        $AllUsers
    }
}
