. $PSScriptRoot/imports.ps1

Describe -Name "Update-Configuration" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Update a configuration file" {
        $Response = PSc8y\Update-Configuration -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a configuration file (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Update-Configuration -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

