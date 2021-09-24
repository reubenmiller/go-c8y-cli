. $PSScriptRoot/imports.ps1

Describe -Name "Get-FirmwareVersion" {
    BeforeEach {
        $mo = PSc8y\New-ManagedObject -Name "testMO"
        $mo = PSc8y\New-FirmwareVersion -Firmware 12345 -Version "1.0.0" -Url "test.com/file.mender"

    }

    It "Get a firmware package" {
        $Response = PSc8y\Get-FirmwareVersion -Firmware 12345 -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a firmware package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Get-FirmwareVersion
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue

    }
}

