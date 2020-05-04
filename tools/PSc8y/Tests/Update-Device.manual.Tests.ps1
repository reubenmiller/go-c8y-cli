. $PSScriptRoot/imports.ps1

Describe -Name "Update-Device" {
    BeforeEach {
        $device = PSc8y\New-TestDevice

    }

    It "Update device by id" {
        $Response = PSc8y\Update-Device -Id $device.id -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.name | Should -BeExactly "MyNewName"
    }

    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

