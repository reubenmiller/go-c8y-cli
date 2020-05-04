. $PSScriptRoot/imports.ps1

Describe -Name "New-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile

    }

    It "Add a binary to an event" {
        $Response = PSc8y\New-EventBinary -Id $Event.id -File $TestFile
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}

