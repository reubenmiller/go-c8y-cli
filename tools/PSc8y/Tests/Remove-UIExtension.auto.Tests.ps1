. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UIExtension" {
    BeforeEach {

    }

    It -Skip "Remove UI extension" {
        $Response = PSc8y\Remove-UIExtension -Id $App.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

