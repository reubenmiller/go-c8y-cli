. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceChildAdditionCollection" {
    BeforeEach {

    }

    It "Get a list of the child additions of an existing device" {
        $Response = PSc8y\Get-DeviceChildAdditionCollection -Device agentAdditionInfo01
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "List child additions of a device but filter the children using a custom query" {
        $Response = PSc8y\"agentAdditionInfo01" | Get-DeviceChildAdditionCollection -Query "type eq 'custom*'"

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

