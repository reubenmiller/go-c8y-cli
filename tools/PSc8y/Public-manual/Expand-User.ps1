Function Expand-User {
<#
.SYNOPSIS
Expand a list of users

.DESCRIPTION
Expand a list of users replacing any ids or names with the actual user object.

.NOTES
If the given object is already an user object, then it is added with no additional lookup

.EXAMPLE
Expand-User "myuser"

Retrieve the user objects by name or id

.EXAMPLE
Get-UserCollection *test* | Expand-User

Get all the user object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

#>
    [cmdletbinding()]
    Param(
        # List of ids, names or user objects
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        [array] $AllUsers = foreach ($iUser in $InputObject)
        {
            if (($iUser -is [string]))
            {
                if ($iUser -match "\*") {
                    # Remove any wildcard characters, as they are not supported by c8y
                    $iUserNormalized = $iUser -replace "\*", ""

                    PSc8y\Get-UserCollection -Username $iUserNormalized -Dry:$false -PageSize 100 -AsPSObject |
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
