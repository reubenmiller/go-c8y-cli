. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceRequestCollection" {
    BeforeEach {
        $DeviceRequest = Register-Device -Id "919293993939393"

    }

    It "Get a list of new device requests" {
        $Response = PSc8y\Get-DeviceRequestCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceRequest -Id "919293993939393"

    }
}

