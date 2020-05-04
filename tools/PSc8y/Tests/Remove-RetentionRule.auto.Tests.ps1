. $PSScriptRoot/imports.ps1

Describe -Name "Remove-RetentionRule" {
    BeforeEach {
        $RetentionRule = New-RetentionRule -DataType ALARM -MaximumAge 200

    }

    It "Delete a retention rule" {
        $Response = PSc8y\Remove-RetentionRule -Id $RetentionRule.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a retention rule (using pipeline)" {
        $Response = PSc8y\Get-RetentionRule -Id $RetentionRule.id | Remove-RetentionRule
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

