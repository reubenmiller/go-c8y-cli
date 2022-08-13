. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildAssetCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of the child assets of an existing device" {
        $Response = PSc8y\Get-DeviceGroupChildAssetCollection -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

