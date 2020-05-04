. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceRequest" {
    BeforeEach {
        $id = "010af8dd0c102"
        $DeviceRequest = Register-Device -Id $id

    }

    It "Get a new device request" {
        $Response = PSc8y\Get-DeviceRequest -Id "010af8dd0c102"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceRequest -Id "010af8dd0c102"

    }
}

