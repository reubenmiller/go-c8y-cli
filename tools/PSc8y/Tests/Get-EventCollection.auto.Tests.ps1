. $PSScriptRoot/imports.ps1

Describe -Name "Get-EventCollection" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = PSc8y\New-Event -Device $Device.id -Type "my_CustomType2" -Time "-9d" -Text "Test event"

    }

    It "Get events with type 'my_CustomType' that were created in the last 10 days" {
        $Response = PSc8y\Get-EventCollection -Type "my_CustomType2" -DateFrom "-10d"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get events from a device" {
        $Response = PSc8y\Get-EventCollection -Device $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get events from a device (using pipeline)" {
        $Response = PSc8y\Get-DeviceCollection -Name $Device.name | Get-EventCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}

