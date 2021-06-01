Function Get-RequestBodyCollection {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true
        )]
        [object[]] $Response
    )

    Begin {
        $inputItems = New-Object System.Collections.ArrayList
    }

    Process {
        $null = $inputItems.AddRange($Response)
    }

    End {
        $EntireResponse = $inputItems | Out-String

        $MultiLineMatches = Select-String -InputObject $EntireResponse -AllMatches -Pattern "(?smi)Body:\s*({.+?\s})\s"
        $SingleLineMatches = Select-String -InputObject $EntireResponse -AllMatches -Pattern "(?mi)Body:\s*({.+?})\s*$"
        $AllMatches = @() + $MultiLineMatches.Matches + $SingleLineMatches.Matches | Sort-Object Index

        $BodyMatches = $AllMatches
        | Where-Object { $_.Groups.Count -gt 1 } `
        | ForEach-Object { ConvertFrom-Json $_.Groups[1] -Depth 100 }

        $BodyMatches
    }
}
