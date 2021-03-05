. $PSScriptRoot/imports.ps1

Describe -Name "Update-CurrentUser" {
    BeforeEach {

    }

    It "Update the current user's last name" {
        $Response = PSc8y\Update-CurrentUser -LastName "Smith"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

