. $PSScriptRoot/imports.ps1

Describe -Name "New-ExternalId" {
    BeforeEach {
        $TestDevice = PSc8y\New-TestDevice

    }

    It "Get external identity" {
        $Response = PSc8y\New-ExternalId -Device $TestDevice.id -Type "my_SerialNumber" -Name "myserialnumber"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestDevice.id) {
            PSc8y\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue
        }

    }
}

