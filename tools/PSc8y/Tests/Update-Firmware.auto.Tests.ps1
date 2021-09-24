. $PSScriptRoot/imports.ps1

Describe -Name "Update-Firmware" {
    BeforeEach {
        $mo = PSc8y\New-Firmware -Name "package1"

    }

    It "Update a firmware package name and add custom add custom properties" {
        $Response = PSc8y\Update-Firmware -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Update-Firmware -Data @{ com_my_props = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

