. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Get service status" {
        $Response = PSc8y\Get-DeviceSoftware -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

