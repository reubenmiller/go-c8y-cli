. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Binary" {
    BeforeEach {
        $File = New-TestFile
        $Binary = New-Binary -File $File

    }

    It "Delete a binary" {
        $Response = PSc8y\Remove-Binary -Id $Binary.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-Item $File

    }
}

