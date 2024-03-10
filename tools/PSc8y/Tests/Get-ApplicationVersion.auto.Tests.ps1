. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationVersion" {
    BeforeEach {

    }

    It "Get application version by tag" {
        $Response = PSc8y\Get-ApplicationVersion -Id 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get application version by version name" {
        $Response = PSc8y\Get-ApplicationVersion -Id 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

