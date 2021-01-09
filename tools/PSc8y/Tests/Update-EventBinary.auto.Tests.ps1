. $PSScriptRoot/imports.ps1

Describe -Name "Update-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id -WithBinary
        $TestFile = New-TestFile

    }

    It "Update a binary related to an event" {
        $Response = PSc8y\Update-EventBinary -Id $Event.id -File $TestFile
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}

