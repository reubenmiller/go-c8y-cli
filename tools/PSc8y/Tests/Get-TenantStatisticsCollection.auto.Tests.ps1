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

    It "Get tenant statistics collection for the day before yesterday" {
        $Response = PSc8y\Get-TenantStatisticsCollection -DateFrom "-3d" -DateTo "-2d"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

