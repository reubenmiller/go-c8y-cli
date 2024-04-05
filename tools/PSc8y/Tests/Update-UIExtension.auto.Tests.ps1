. $PSScriptRoot/imports.ps1

Describe -Name "Update-UIExtension" {
    BeforeEach {

    }

    It -Skip "Update application availability to MARKET" {
        $Response = PSc8y\Update-UIExtension -Id $App.name -Availability "MARKET"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

