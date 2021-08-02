. $PSScriptRoot/imports.ps1

Describe -Name "Get-ConfigurationCollection" {
    BeforeEach {

    }

    It "Get a list of configuration files" {
        $Response = PSc8y\Get-ConfigurationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

