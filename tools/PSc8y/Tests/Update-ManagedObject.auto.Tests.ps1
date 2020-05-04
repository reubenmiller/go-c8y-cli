. $PSScriptRoot/imports.ps1

Describe -Name "Update-ManagedObject" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Update a managed object" {
        $Response = PSc8y\Update-ManagedObject -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a managed object (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Update-ManagedObject -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

