. $PSScriptRoot/imports.ps1

Describe -Name "Update-RetentionRule" {
    BeforeEach {
        $RetentionRule = New-RetentionRule -DataType ALARM -MaximumAge 365

    }

    It "Update a retention rule" {
        $Response = PSc8y\Update-RetentionRule -Id $RetentionRule.id -DataType MEASUREMENT -FragmentType "custom_FragmentType"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a retention rule (using pipeline)" {
        $Response = PSc8y\Get-RetentionRule -Id $RetentionRule.id | Update-RetentionRule -DataType MEASUREMENT -FragmentType "custom_FragmentType"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-RetentionRule -Id $RetentionRule.id

    }
}

