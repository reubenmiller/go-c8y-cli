. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceCertificate" {
    BeforeEach {

    }

    It "Update device certificate by id/fingerprint" {
        $Response = PSc8y\Update-DeviceCertificate -Id abcedef0123456789abcedef0123456789 -Status DISABLED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update device certificate by name" {
        $Response = PSc8y\Update-DeviceCertificate -Id "MyCert" -Status DISABLED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

