Function Expand-Source {
<#
.SYNOPSIS
Expand a list of source ids.

.DESCRIPTION
1. Look for a source.id property
2. Look for a deviceId property
3. Look for a id property
4. Check if the given is a string or int and is integer like

.PARAMETER InputObject
List of ids

.EXAMPLE
Expand-Source 12345

Normalize a list of ids

.EXAMPLE
"12345", "56789" | Expand-Source

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
        [array] $AllIds = foreach ($iObject in $InputObject)
        {
            $currentID = $iObject

            if ($null -ne $iObject.source.id) {
                $currentID = $iObject.source.id
            } elseif ($null -ne $iObject.deviceId) {
                $currentID = $iObject.deviceId
            } elseif ($null -ne $iObject.id) {
                $currentID = $iObject.id
            }

            # Allow for matching integer or strings types, hence the the quotes around the $currentID variable
            if ("$currentID" -match "^[0-9]+$")
            {
                $currentID
            }
        }
        $AllIds
    }
}
