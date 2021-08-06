. $PSScriptRoot/imports.ps1

Describe -Name "Remove-FirmwareVersionPatch" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

    }

    It "Delete a firmware package version patch" {
        $Response = PSc8y\Remove-FirmwareVersionPatch -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package version patch (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-FirmwareVersionPatch
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package version patch and related binary" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-FirmwareVersionPatch -ForceCascade
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $ChildDevice.id -ErrorAction SilentlyContinue

    }
}

