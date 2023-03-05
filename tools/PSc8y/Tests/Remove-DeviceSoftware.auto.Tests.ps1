. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Remove software" {
        $Response = PSc8y\Remove-DeviceSoftware -Id 12345 -Name ntp
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

