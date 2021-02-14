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
        $BodyMatches = $inputItems `
        | Out-String `
        | Select-String -AllMatches -Pattern "(?smi)Body:\s*({.+?}[^,])" `
        | ForEach-Object { $_.Matches } `
        | Where-Object { $_.Groups.Count -gt 1 } `
        | ForEach-Object { ConvertFrom-Json $_.Groups[1] -Depth 100 }

        $BodyMatches
    }
}
