. $PSScriptRoot/imports.ps1

Describe -Name "Remove-SoftwareVersion" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

    }

    It "Delete a software package" {
        $Response = PSc8y\Remove-Software -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a software package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-Software
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a software package and all related versions" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-Software -Cascade
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $ChildDevice.id -ErrorAction SilentlyContinue

    }
}

