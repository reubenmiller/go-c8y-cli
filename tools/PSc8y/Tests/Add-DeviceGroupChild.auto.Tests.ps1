﻿. $PSScriptRoot/imports.ps1

Describe -Name "Add-DeviceGroupChild" {
    BeforeEach {

    }

    It -Skip "Add a related managed object as a child addition to an existing managed object" {
        $Response = PSc8y\Add-DeviceGroupChild -Id $software.id -Child $version.id -ChildType addition
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

