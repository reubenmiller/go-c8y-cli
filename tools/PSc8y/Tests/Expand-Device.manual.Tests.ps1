﻿. $PSScriptRoot/imports.ps1

Describe -Name "Expand-Device" {
    BeforeAll {
        $Device = PSc8y\New-TestAgent
    }

    It "Expand device (with object)" {
        $Result = PSc8y\Expand-Device $Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with object) by pipeline" {
        $Result = $Device | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with id)" {
        $Result = PSc8y\Expand-Device $Device.id
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with id) by pipeline" {
        $Result = $Device.id | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with name)" {
        $Result = PSc8y\Expand-Device $Device.name
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device (with name) by pipeline" {
        $Result = $Device.name | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from Get-DeviceCollection" {
        $Result = Get-DeviceCollection $Device.name | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from operation" {
        $operation = PSc8y\New-TestOperation -Device $Device.id
        $Result = $operation | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    It "Expand device from operation" {
        $measurement = PSc8y\New-TestMeasurement -Device $Device.id
        $Result = $measurement | PSc8y\Expand-Device
        $Result.id | Should -BeExactly $Device.id
    }

    AfterAll {
        Remove-ManagedObject -Id $Device.id
    }
}
