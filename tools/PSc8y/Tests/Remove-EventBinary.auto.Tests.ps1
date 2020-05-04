. $PSScriptRoot/imports.ps1

Describe -Name "Remove-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile
        New-EventBinary -Id $Event.id -File $TestFile

    }

    It "Delete an binary attached to an event" {
        $Response = PSc8y\Remove-EventBinary -Id $Event.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}

