. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildAddition" {
    BeforeEach {

    }

    It -Skip "Get an existing child addition reference" {
        $Response = PSc8y\Get-DeviceGroupChildAddition -Group $Agent.id -Child $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

