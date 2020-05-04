. $PSScriptRoot/imports.ps1

Describe -Name "New-Application" {
    BeforeEach {

    }

    It "Create new hosted application" {
        $Response = PSc8y\New-Application -Name myapp -Type HOSTED -Key "myapp-key" -ContextPath "myapp"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id "myapp"

    }
}

