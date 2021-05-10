. $PSScriptRoot/imports.ps1

Describe -Name "Get-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile
        New-EventBinary -Id $Event.id -File $TestFile
        Remove-Item $TestFile

    }

    It "Download a binary related to an event" {
        $Response = PSc8y\Get-EventBinary -Id $Event.id -OutputFileRaw ./eventbinary.txt
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Item "./eventbinary.txt"
        Remove-ManagedObject -Id $Device.id

    }
}

