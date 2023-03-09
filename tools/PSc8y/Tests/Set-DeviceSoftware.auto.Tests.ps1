. $PSScriptRoot/imports.ps1

Describe -Name "Set-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Set/replace the list of software on a device" {
        $Response = PSc8y\Set-DeviceSoftware -Device $software.id -Name ntp -Version 1.0.2 -Type apt
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

