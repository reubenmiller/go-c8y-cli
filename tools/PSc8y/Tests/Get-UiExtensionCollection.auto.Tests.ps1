﻿. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIExtensionCollection" {
    BeforeEach {

    }

    It "Get ui extensions" {
        $Response = PSc8y\Get-UIExtensionCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}
