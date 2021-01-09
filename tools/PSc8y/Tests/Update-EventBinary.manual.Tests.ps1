. $PSScriptRoot/imports.ps1

Describe -Name "Update-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile
    }

    It "Create and update a binary attached to an event" {
        $ExpectedText1 = "my test file`nsecond line"
        $ExpectedText1 | Out-File $TestFile
        PSc8y\New-EventBinary -Id $Event.id -File $TestFile
        $BinaryContents = (PSc8y\Get-EventBinary -Id $Event.id) -join "`n"
        $BinaryContents | Should -BeExactly $ExpectedText1

        $ExpectedText2 = "adjusted text äöüß"
        $ExpectedText2 | Out-File $TestFile
        $BinaryContents = PSc8y\Update-EventBinary -Id $Event.id -File $TestFile
        $BinaryContents = (PSc8y\Get-EventBinary -Id $Event.id) -join "`n"
        $BinaryContents | Should -BeExactly $ExpectedText2
    }


    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}

