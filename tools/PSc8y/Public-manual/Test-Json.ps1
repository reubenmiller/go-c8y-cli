Function Test-Json {
<# 
.SYNOPSIS
Test if the input object is a valid json string

.DESCRIPTION
Test the given input to check if it is most likely valid json. The cmdlet uses
a quick json sanity check rather than trying to parse the json to save time.

.EXAMPLE
Test-Json '{ "name": "tester" }'

Returns true if the input data is valid json

#>
    [cmdletbinding()]
    [OutputType([bool])]
    Param(
        # Input data
        [Parameter(
            Mandatory = $true,
            Position = 0,
            ValueFromPipeline = $true
        )]
        [object]
        $InputObject
    )

    Process {
        if ($inputObject -isnot [string]) {
            return $false
        }

        if (!(($InputObject -match "^\s*[\[\{]") -and ($InputObject -match "[\]\}]\s*$"))) {
            Write-Information "Only json array or objects are supported"
            return $false
        }

        $IsValid = $false
        try {
            $null = ConvertFrom-Json -InputObject $InputObject -ErrorAction Stop
            $IsValid = $true
        } catch {
            Write-Information "Invalid json: $_"
        }
        $IsValid
    }
}
