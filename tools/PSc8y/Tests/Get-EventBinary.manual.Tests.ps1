. $PSScriptRoot/imports.ps1

Describe -Name "Get-EventBinary" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id
        $TestFile = New-TestFile
        New-EventBinary -Id $Event.id -File $TestFile

    }

    It "Download a binary related to an event should have expected contents" {
        $tmpfile = New-TemporaryFile
        $Response = PSc8y\Get-EventBinary -Id $Event.id -OutputFileRaw $tmpfile
        $LASTEXITCODE | Should -Be 0
        $tmpfile | Should -Exist
        (Get-Content $tmpfile -Raw) | Should -BeExactly (Get-Content -Raw -LiteralPath $TestFile)
        $Response | Should -BeExactly (Get-Content -Raw -LiteralPath $TestFile).TrimEnd()
        Remove-Item $tmpfile
    }

    It "Download a binary related to an event should have expected contents (using pipeline)" {
        $tmpfile = New-TemporaryFile
        $Response = $Event | PSc8y\Get-EventBinary -OutputFileRaw $tmpfile
        $LASTEXITCODE | Should -Be 0
        $tmpfile | Should -Exist
        (Get-Content $tmpfile -Raw) | Should -BeExactly (Get-Content -Raw -LiteralPath $TestFile)
        $Response | Should -BeExactly (Get-Content -Raw -LiteralPath $TestFile).TrimEnd()
        Remove-Item $tmpfile
    }

    AfterEach {
        Remove-Item $TestFile
        Remove-ManagedObject -Id $Device.id

    }
}
