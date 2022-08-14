. $PSScriptRoot/imports.ps1

Describe -Name "Remove-FirmwarePatch" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $Device = PSc8y\New-TestDevice
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\Add-ManagedObjectChild -ChildType device -Id $Device.id -Child $ChildDevice.id

    }

    It "Delete a firmware package version patch" {
        $Response = PSc8y\Remove-FirmwarePatch -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware patch (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-FirmwarePatch
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware patch and related binary" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-FirmwarePatch -ForceCascade
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $ChildDevice.id -ErrorAction SilentlyContinue

    }
}

