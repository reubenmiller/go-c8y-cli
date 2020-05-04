. $PSScriptRoot/imports.ps1

Describe -Name "Copy-Application" {
    BeforeEach {
        New-Application -Name my-example-app -Type HOSTED -Key "my-example-app-key" -ContextPath "my-example-app"

    }

    It "Copy an existing application" {
        $Response = PSc8y\Copy-Application -Id "my-example-app"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id "my-example-app"
        Remove-Application -Id "clonemy-example-app"

    }
}

