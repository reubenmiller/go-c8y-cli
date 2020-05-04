. $PSScriptRoot/imports.ps1

Describe -Name "New-Operation" {
    BeforeEach {
        $device = New-TestAgent

    }

    It "Create operation for a device" {
        $Response = PSc8y\New-Operation -Device $device.id -Description "Restart device" -Data @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create operation for a device (using pipeline)" {
        $Response = PSc8y\Get-Device $device.id | New-Operation -Description "Restart device" -Data @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

