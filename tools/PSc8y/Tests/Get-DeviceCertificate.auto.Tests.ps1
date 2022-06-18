. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceCertificate" {
    BeforeEach {

    }

    It "Get trusted device certificate by id/fingerprint" {
        $Response = PSc8y\Get-DeviceCertificate -Id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

