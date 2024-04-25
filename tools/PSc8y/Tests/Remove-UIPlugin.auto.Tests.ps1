. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UIPlugin" {
    BeforeEach {

    }

    It -Skip "Remove UI plugin" {
        $Response = PSc8y\Remove-UIPlugin -Id $App.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

