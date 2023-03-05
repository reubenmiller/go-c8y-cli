. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Update software status" {
        $Response = PSc8y\Update-DeviceSoftware -Id 12345 -Status up
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

