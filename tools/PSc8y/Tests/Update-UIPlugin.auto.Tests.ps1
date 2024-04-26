. $PSScriptRoot/imports.ps1

Describe -Name "Update-UIPlugin" {
    BeforeEach {

    }

    It -Skip "Update plugin availability to MARKET" {
        $Response = PSc8y\Update-UIPlugin -Id $App.name -Availability "MARKET"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

