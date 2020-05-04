. $PSScriptRoot/imports.ps1

Describe -Name "Update-Binary" {
    BeforeEach {
        $File1 = New-TestFile
        $File2 = New-TestFile
        $Binary1 = New-Binary -File $File1
        $FileName1 = (Get-Item $File1).Name

    }

    It "Update an existing binary file" {
        $Response = PSc8y\Update-Binary -Id $Binary1.id -File $File2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Item $File1
        Remove-Item $File2
        Find-ManagedObjectCollection -Query "has(c8y_IsBinary) and (name eq '$FileName1')" | Remove-Binary

    }
}

