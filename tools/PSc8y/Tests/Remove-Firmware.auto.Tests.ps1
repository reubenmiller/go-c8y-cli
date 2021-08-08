. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Firmware" {
    BeforeEach {
        $mo = PSc8y\New-Firmware -Name "firmware1"

    }

    It "Delete a firmware package and all related versions" {
        $Response = PSc8y\Remove-Firmware -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-Firmware
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package but keep the binaries" {
        $Response = PSc8y\Get-ManagedObject -Id $Device.id | Remove-Firmware -ForceCascade:$false
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue
        Remove-ManagedObject -Id $Device.id -ErrorAction SilentlyContinue

    }
}

