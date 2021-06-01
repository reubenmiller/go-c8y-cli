. $PSScriptRoot/imports.ps1

Describe -Name "Get-ManagedObjectCollection" {
    BeforeEach {
        $Device1 = New-TestDevice
        $Device2 = New-TestDevice

    }

    It "Get a list of managed objects" {
        $Response = PSc8y\Get-ManagedObjectCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of managed objects by id" {
        $Response = PSc8y\Get-ManagedObjectCollection -Ids $Device1.id, $Device2.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device1.id
        Remove-ManagedObject -Id $Device2.id

    }
}

