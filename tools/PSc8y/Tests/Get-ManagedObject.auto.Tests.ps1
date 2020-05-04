. $PSScriptRoot/imports.ps1

Describe -Name "Get-ManagedObject" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Get a managed object" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a managed object (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-ManagedObject
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a managed object with parent references" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id -WithParents
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

