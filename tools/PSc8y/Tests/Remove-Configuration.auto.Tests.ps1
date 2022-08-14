. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Configuration" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ManagedObjectChild -ChildType childDevices -Id $Device.id -Child $ChildDevice.id

    }

    It "Delete a configuration package (and any related binaries)" {
        $Response = PSc8y\Remove-Configuration -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a configuration package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-Configuration
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a configuration package but keep any related binaries" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-Configuration -forceCascade:$false
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $ChildDevice.id -ErrorAction SilentlyContinue

    }
}

