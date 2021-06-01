. $PSScriptRoot/imports.ps1

Describe -Name "Get-InventoryRole" {
    BeforeEach {

    }

    It "Get an inventory role (using pipeline)" {
        $Response = PSc8y\Get-InventoryRoleCollection -PageSize 1 | Get-InventoryRole
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

