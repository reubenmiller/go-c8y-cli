. $PSScriptRoot/imports.ps1

Describe -Name "Get-Software" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Get a software package" {
        $Response = PSc8y\Get-Software -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a software package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-Software
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id

    }
}

