. $PSScriptRoot/imports.ps1

Describe -Name "Remove-OperationCollection" {
    BeforeEach {
        $TestDevice = PSc8y\New-TestDevice

    }

    It "Remove all pending operations for a given device" {
        $Response = PSc8y\Remove-OperationCollection -Device $TestDevice.id -Status PENDING
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        if ($TestDevice.id) {
            PSc8y\Remove-ManagedObject -Id $TestDevice.id -ErrorAction SilentlyContinue
        }

    }
}

