. $PSScriptRoot/imports.ps1

Describe -Name "Get-TenantStatisticsCollection" {
    BeforeEach {

    }

    It "Get tenant statistics collection" {
        $Response = PSc8y\Get-TenantStatisticsCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get tenant statistics collection for the last 30 days" {
        $Response = PSc8y\Get-TenantStatisticsCollection -DateFrom "-30d" -PageSize 30
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get tenant statistics collection for the last 10 days, only return until the last 9 days" {
        $Response = PSc8y\Get-TenantStatisticsCollection -DateFrom "-10d" -DateTo "-9d"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

