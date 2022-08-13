. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildAsset" {
    BeforeEach {

    }

    It -Skip "Get an existing child asset reference" {
        $Response = PSc8y\Get-DeviceGroupChildAsset -Group $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

