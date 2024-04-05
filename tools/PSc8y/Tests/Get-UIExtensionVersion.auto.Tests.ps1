. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIExtensionVersion" {
    BeforeEach {

    }

    It "Get extension version by tag" {
        $Response = PSc8y\Get-UIExtensionVersion -Extension 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get extension version by version name" {
        $Response = PSc8y\Get-UIExtensionVersion -Extension 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

