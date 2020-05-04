. $PSScriptRoot/imports.ps1

Describe -Name "Get-Application" {
    BeforeEach {
        $App = New-Application -Name my-simple-app -Type HOSTED -Key "my-simple-app-key" -ContextPath "my-simple-app"

    }

    It "Get an application by id" {
        $Response = PSc8y\Get-Application -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get an application by name" {
        $Response = PSc8y\Get-Application -Id "my-simple-app"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id "my-simple-app"

    }
}

