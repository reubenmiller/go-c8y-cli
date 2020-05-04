. $PSScriptRoot/imports.ps1

Describe -Name "Get-RetentionRule" {
    BeforeEach {
        $RetentionRule = New-RetentionRule -DataType ALARM -MaximumAge 365

    }

    It "Get a retention rule" {
        $Response = PSc8y\Get-RetentionRule -Id $RetentionRule.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-RetentionRule -Id $RetentionRule.id

    }
}

