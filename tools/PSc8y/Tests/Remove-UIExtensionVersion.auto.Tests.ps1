. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UIExtensionVersion" {
    BeforeEach {

    }

    It "Delete extension version by tag" {
        $Response = PSc8y\Remove-UIExtensionVersion -Extension 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete extension version by version name" {
        $Response = PSc8y\Remove-UIExtensionVersion -Extension 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

