. $PSScriptRoot/imports.ps1

Describe -Name "Get-Binary" {
    BeforeEach {
        $File = New-TestFile
        $Binary = PSc8y\New-Binary -File $File

    }

    It "Get a binary and display the contents on the console" {
        $Response = PSc8y\Get-Binary -Id $Binary.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a binary and save it to a file" {
        $Response = PSc8y\Get-Binary -Id $Binary.id -OutputFileRaw ./download-binary1.txt
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-Binary -Id $Binary.id
        if (Test-Path "./download-binary1.txt") { Remove-Item ./download-binary1.txt }
        Remove-Item $File

    }
}

