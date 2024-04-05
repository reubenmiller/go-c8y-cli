. $PSScriptRoot/imports.ps1

Describe -Name "Update-UIExtensionVersion" {
    BeforeEach {

    }

    It "Replace tags assigned to a version of an extension" {
        $Response = PSc8y\Update-UIExtensionVersion -Extension 1234 -Version 1.0 -Tag tag1,latest
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

