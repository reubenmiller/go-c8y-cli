. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceProfileCollection" {
    BeforeEach {

    }

    It "Get a list of device profiles" {
        $Response = PSc8y\Get-DeviceProfileCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

