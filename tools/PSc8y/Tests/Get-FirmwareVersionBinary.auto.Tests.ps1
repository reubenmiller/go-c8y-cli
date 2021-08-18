. $PSScriptRoot/imports.ps1

Describe -Name "Get-FirmwareVersionBinary" {
    BeforeEach {

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

    }
}

