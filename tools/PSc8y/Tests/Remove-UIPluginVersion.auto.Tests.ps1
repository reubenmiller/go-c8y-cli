. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UIPluginVersion" {
    BeforeEach {

    }

    It -Skip "Delete plugin version by tag" {
        $Response = PSc8y\Remove-UIPluginVersion -Plugin 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete plugin version by version name" {
        $Response = PSc8y\Remove-UIPluginVersion -Plugin 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

