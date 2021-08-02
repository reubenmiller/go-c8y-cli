. $PSScriptRoot/imports.ps1

Describe -Name "Get-Firmware" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Get a firmware package" {
        $Response = PSc8y\Get-Firmware -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-Firmware
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

