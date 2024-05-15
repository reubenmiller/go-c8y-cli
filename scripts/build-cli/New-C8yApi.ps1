Function New-C8yApi
 {
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

            if ([string]::IsNullOrWhiteSpace($Specification.group.name)) {
                Write-Warning "Skipping spec: Specification is missing the information.name property. This is required. file=$Path"
                continue
            }

            if ($Specification.group.skip -eq $true) {
                Write-Warning "Skipping spec: Specification is marked to be skipped using information.skip property. This is required. file=$Path"
                continue
            }

            # Create root command (golang convention is to use lower case packages names)
            $packageName = $Specification.group.name.ToLower()
            $CommandOutput = Join-Path $OutputDir -ChildPath $packageName
            $null = New-Item -Path $CommandOutput -ItemType Directory -Force
            New-C8yApiGoRootCommand -Specification:$Specification -OutputDir:$CommandOutput
            Write-Host ("Generating api root command: {0}" -f $CommandOutput) -ForegroundColor Cyan
			foreach ($iSpec in $Specification.commands) {
                if ($iSpec.skip -eq $true) {
                    Write-Verbose ("Skipping [{0}]" -f $iSpec.name)
                    continue
                }
                $SubCommandPackageName = $iSpec.alias.go.ToLower() -replace "-", "_"
                $SubCommandOutput = Join-Path -Path $CommandOutput -ChildPath $SubCommandPackageName
				$null = New-Item -Path $SubCommandOutput -ItemType Directory -Force
                Write-Host ("Generating subcommand: {0}" -f $SubCommandOutput) -ForegroundColor Magenta
                New-C8yApiGoCommand -Specification:$iSpec -OutputDir:$SubCommandOutput -ParentName $packageName
			}
        }

        $importStatements
    }
}
