. $PSScriptRoot/imports.ps1

Describe -Name "Get-BinaryCollection" {
    BeforeEach {
        $File = New-TestFile
        $Binary = PSc8y\New-Binary -File $File

    }

    It "Get a list of binaries" {
        $Response = PSc8y\Get-BinaryCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-Binary -Id $Binary.id

    }
}

