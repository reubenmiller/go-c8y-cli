. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentUserInventoryRole" {
    BeforeEach {

    }

    It "Get an inventory role of the current user (using pipeline)" {
        $Response = PSc8y\Get-CurrentUserInventoryRoleCollection -PageSize 1 | Get-CurrentUserInventoryRole
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

