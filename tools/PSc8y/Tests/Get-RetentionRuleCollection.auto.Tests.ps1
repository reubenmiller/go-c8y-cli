. $PSScriptRoot/imports.ps1

Describe -Name "Get-RetentionRuleCollection" {
    BeforeEach {

    }

    It "Get a list of retention rules" {
        $Response = PSc8y\Get-RetentionRuleCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

