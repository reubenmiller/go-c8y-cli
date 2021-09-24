. $PSScriptRoot/imports.ps1

Describe -Name "Get-Configuration" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Get a configuration package" {
        $Response = PSc8y\Get-Configuration -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a configuration package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-Configuration
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

