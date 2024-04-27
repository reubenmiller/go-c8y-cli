. $PSScriptRoot/imports.ps1

Describe -Name "Get-UIPlugin" {
    BeforeEach {

    }

    It -Skip "Get ui plugin" {
        $Response = PSc8y\Get-UIPlugin -Id 1234
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

