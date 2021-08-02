. $PSScriptRoot/imports.ps1

Describe -Name "Update-Software" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Update a software package" {
        $Response = PSc8y\Update-Software -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a software package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Update-Software -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

