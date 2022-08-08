. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceChildAssetCollection" {
    BeforeEach {

    }

    It "Get a list of the child assets of an existing device" {
        $Response = PSc8y\Get-DeviceChildAssetCollection -Device agentAssetInfo01
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "List child assets of a device but filter the children using a custom query" {
        $Response = PSc8y\"agentAssetInfo01" | Get-DeviceChildAssetCollection -Query "type eq 'custom*'"

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

