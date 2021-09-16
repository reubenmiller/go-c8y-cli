﻿. $PSScriptRoot/imports.ps1

Describe -Name "Install-SoftwareVersion" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $mo = PSc8y\New-SoftwareVersion -SoftwareId 12345 -Version "1.0.0" -Url "test.com/file.mender"

    }

    It "Get a software package" {
        $Response = PSc8y\Get-SoftwareVersion -SoftwareId 12345 -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a software package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-SoftwareVersion
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue

    }
}

