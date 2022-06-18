. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceCertificate" {
    BeforeEach {

    }

    It "Remove trusted device certificate by id/fingerprint" {
        $Response = PSc8y\Remove-DeviceCertificate -Id abcedef0123456789abcedef0123456789
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

