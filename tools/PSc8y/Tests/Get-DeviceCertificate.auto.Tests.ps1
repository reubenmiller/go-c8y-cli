. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceCertificate" {
    BeforeEach {

    }

    It -Skip "Get trusted device certificate by id/fingerprint" {
        $Response = PSc8y\Get-DeviceCertificate -Id abcedef0123456789abcedef0123456789
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

