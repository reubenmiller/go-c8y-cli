. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentUserInventoryRoleCollection" {
    BeforeEach {

    }

    It "Get the current users inventory roles" {
        $Response = PSc8y\Get-CurrentUserInventoryRoleCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

