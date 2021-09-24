. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Firmware" {
    BeforeEach {
        $name = New-RandomString -Prefix "firmware1"
        $firmware = PSc8y\New-Firmware -Name $name

    }

    It "Delete a firmware package and all related versions" {
        $Response = PSc8y\Remove-Firmware -Id $firmware.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $firmware.id | Remove-Firmware
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package but keep the binaries" {
        $Response = PSc8y\Get-ManagedObject -Id $firmware.id | Remove-Firmware -ForceCascade:$false
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

