. $PSScriptRoot/imports.ps1

Describe -Name "Find-ManagedObjectCollection" {
    BeforeEach {
        $Device = New-TestDevice -Name "roomUpperFloor_"

    }

    It "Find all devices with their names starting with 'roomUpperFloor_'" {
        $Response = PSc8y\Find-ManagedObjectCollection -Query "name eq 'roomUpperFloor_*'"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

