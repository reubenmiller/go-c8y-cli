. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Application" {
    BeforeEach {
        $App = New-Application -Name my-temp-app -Type HOSTED -Key "my-temp-app-key" -ContextPath "my-temp-app"

    }

    It "Delete an application by id" {
        $Response = PSc8y\Remove-Application -Id $App.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete an application by name" {
        $Response = PSc8y\Remove-Application -Id "my-temp-app"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

