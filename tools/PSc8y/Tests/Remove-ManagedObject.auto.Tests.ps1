. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ManagedObject" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

    }

    It "Delete a managed object" {
        $Response = PSc8y\Remove-ManagedObject -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a managed object (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-ManagedObject
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a managed object and all child devices" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-ManagedObject -Cascade
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $ChildDevice.id -ErrorAction SilentlyContinue

    }
}

