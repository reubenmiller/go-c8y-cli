. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceCertificate" {
    BeforeEach {

    }

    It -Skip "Upload a trusted device certificate" {
        $Response = PSc8y\New-DeviceCertificate -Name "MyCert" -File "./cert.pem"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

