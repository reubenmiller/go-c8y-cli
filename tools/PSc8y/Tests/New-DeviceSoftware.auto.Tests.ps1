. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Create a new software for a device" {
        $Response = PSc8y\New-DeviceSoftware -Id $software.id -Name ntp -Version 1.0.2 -Type apt
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

