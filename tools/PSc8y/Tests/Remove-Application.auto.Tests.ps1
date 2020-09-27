. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Application" {
    BeforeEach {
        $App = New-TestHostedApplication

    }

    It "Delete an application by id" {
        $Response = PSc8y\Remove-Application -Id $App.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete an application by name" {
        $Response = PSc8y\Remove-Application -Id $App.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

