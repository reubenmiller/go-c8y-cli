. $PSScriptRoot/imports.ps1

Describe -Name "Get-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile
        New-EventBinary -Id $Event.id -File $TestFile

    }

    It "Download a binary related to an event should have expected contents" {
        $Response = PSc8y\Get-EventBinary -Id $Event.id -OutputFile "./value1.output.txt"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Exist
        $Response | Should -FileContentMatchMultiline (Get-Content -Raw -LiteralPath $TestFile)
        Remove-Item $Response
    }

    It "Download a binary related to an event should have expected contents (using pipeline)" {
        $Response = $Event | PSc8y\Get-EventBinary -OutputFile "./value2.output.txt"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Exist
        $Response | Should -FileContentMatchMultiline (Get-Content -Raw -LiteralPath $TestFile)
        Remove-Item $Response
    }

    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}
