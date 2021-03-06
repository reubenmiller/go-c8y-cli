﻿Function New-C8yApi {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
		[string[]] $InFile,

		[Parameter(
            Mandatory = $true,
            Position = 1
        )]
		[string] $OutputDir
	)

	Begin {
		if (!(Test-Path $OutputDir)) {
			$null = New-Item -Type Directory -Path $OutputDir
		}
	}

    Process {
        $importStatements = foreach ($iFile in $InFile) {
			$Path = Resolve-Path $iFile

            $Specification = Get-Content $Path -Raw -Encoding utf8 | ConvertFrom-Json

            # Create root command
            New-C8yApiGoRootCommand -Specification:$Specification -OutputDir:$OutputDir

			foreach ($iSpec in $Specification.endpoints) {
                if ($iSpec.skip -eq $true) {
                    Write-Verbose ("Skipping [{0}]" -f $iSpec.name)
                    continue
                }
				New-C8yApiGoCommand -Specification:$iSpec -OutputDir:$OutputDir
			}
        }

        $importStatements
    }
}
