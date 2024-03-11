. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ApplicationVersion" {
    BeforeEach {

    }

    It "Delete application version by tag" {
        $Response = PSc8y\Remove-ApplicationVersion -Application 1234 -Tag tag1
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete application version by version name" {
        $Response = PSc8y\Remove-ApplicationVersion -Application 1234 -Version 1.0
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

