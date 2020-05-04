. $PSScriptRoot/imports.ps1

Describe -Name "Remove-EventCollection" {
    BeforeEach {
        $TestDevice = PSc8y\New-TestDevice

    }

    It "Remove events with type 'my_CustomType' that were created in the last 10 days" {
        $Response = PSc8y\Remove-EventCollection -Type my_CustomType -DateFrom "-10d"
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove events from a device" {
        $Response = PSc8y\Remove-EventCollection -Device $TestDevice.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        if ($TestDevice.id) {
            PSc8y\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue
        }

    }
}

