. $PSScriptRoot/imports.ps1

Describe -Name "New-RetentionRule" {
    BeforeEach {

    }

    It "Create a retention rule to delete all alarms after 180 days" {
        $Response = PSc8y\New-RetentionRule -DataType ALARM -MaximumAge 180
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-RetentionRuleCollection -PageSize 100 | Select-Object -Last 1 | Remove-RetentionRule

    }
}

