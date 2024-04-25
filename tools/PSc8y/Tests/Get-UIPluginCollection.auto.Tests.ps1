. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIPluginCollection" {
    BeforeEach {

    }

    It "Get UI plugins" {
        $Response = PSc8y\Get-UIPluginCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

