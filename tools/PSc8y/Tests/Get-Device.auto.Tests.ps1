. $PSScriptRoot/imports.ps1

Describe -Name "Get-Device" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Get device by id" {
        $Response = PSc8y\Get-Device -Id $device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get device by name" {
        $Response = PSc8y\Get-Device -Id $device.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

