. $PSScriptRoot/imports.ps1

Describe -Name "Get-InventoryRoleCollection" {
    BeforeEach {

    }

    It "Get list of inventory roles" {
        $Response = PSc8y\Get-InventoryRoleCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

