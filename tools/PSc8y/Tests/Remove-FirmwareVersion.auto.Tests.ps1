. $PSScriptRoot/imports.ps1

Describe -Name "Remove-FirmwareVersion" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Delete a firmware version and all related versions" {
        $Response = PSc8y\Remove-FirmwareVersion -Id $mo.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-FirmwareVersion
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue

    }
}

