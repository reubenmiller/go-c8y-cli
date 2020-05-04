Function Expand-Id {
<#
.SYNOPSIS
Expand a list of ids.

.PARAMETER InputObject
List of ids

.EXAMPLE
Expand-Id 12345

Normalize a list of ids

.EXAMPLE
"12345", "56789" | Expand-Id

Normalize a list of ids

#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [AllowEmptyCollection()]
        [AllowNull()]
        [object[]] $InputObject
    )

    Process {
        [array] $AllIds = foreach ($iID in $InputObject)
        {
            $currentID = $iID
            if ($null -ne $iID.id) {
                $currentID = $iID.id
            }
            # Allow for matching integer or strings types, hence the the quotes around the $currentID variable
            if ("$currentID" -match "^[0-9a-z_\-*]+$")
            {
                $currentID
            }
        }
        $AllIds
    }
}
