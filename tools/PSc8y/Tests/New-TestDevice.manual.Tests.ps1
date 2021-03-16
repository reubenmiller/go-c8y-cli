. $PSScriptRoot/imports.ps1

Describe -Name "New-TestDevice" {

    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    It "Creates a test device" {
        $output = New-TestDevice
        $null = $ids.Add($output.id)
        $LASTEXITCODE | Should -Be 0
        $output.Name | Should -BeLike "testdevice*"
    }

    AfterEach {
        $ids | c8y devices delete
    }
}
