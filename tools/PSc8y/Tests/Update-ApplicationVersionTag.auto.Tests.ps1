﻿. $PSScriptRoot/imports.ps1

Describe -Name "Update-ApplicationVersionTag" {
    BeforeEach {

    }

    It -Skip "Get application version by tag" {
        $Response = PSc8y\Update-ApplicationVersionTag -Application 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get application version by version name" {
        $Response = PSc8y\Update-ApplicationVersionTag -Application 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

