. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceStatisticsCollection" {
    BeforeEach {

    }

    It "Get device statistics" {
        $Response = PSc8y\Get-DeviceStatisticsCollection
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

