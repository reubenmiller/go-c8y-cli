﻿. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildCollection" {
    BeforeEach {

    }

    It "Get a list of the child additions of an existing managed object" {
        $Response = PSc8y\Get-DeviceGroupChildCollection -Id 12345 -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of the child additions of an existing managed object (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id 12345 | Get-DeviceGroupChildCollection -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

