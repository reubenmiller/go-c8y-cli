. $PSScriptRoot/imports.ps1

Describe -Name "Update-Device" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Update device by id" {
        $Response = PSc8y\Update-Device -Id $device.id -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update device by name" {
        $Response = PSc8y\Update-Device -Id $device.name -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update device custom properties" {
        $Response = PSc8y\Update-Device -Id $device.name -Data @{ "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

