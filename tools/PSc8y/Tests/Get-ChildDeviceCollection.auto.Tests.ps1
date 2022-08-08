. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildDeviceCollection" {
    BeforeEach {

    }

    It "Get a list of the child devices of an existing device" {
        $Response = PSc8y\Get-ChildDeviceCollection -Device agentParent01
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of child devices which a specific type" {
        $Response = PSc8y\"agentParent01" | Get-ChildDeviceCollection -Query "type eq 'custom*'"

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

