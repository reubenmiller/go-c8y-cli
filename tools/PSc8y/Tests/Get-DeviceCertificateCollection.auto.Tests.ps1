. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceCertificateCollection" {
    BeforeEach {

    }

    It "Get list of trusted device certificates" {
        $Response = PSc8y\Get-DeviceCertificateCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

