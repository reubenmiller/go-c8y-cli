. $PSScriptRoot/imports.ps1

Describe -Name "Approve-DeviceRequest" {
    BeforeEach {
        $DeviceRequest = Register-Device -Id "1234010101s01ldk208"
        $DeviceCreds = Request-DeviceCredentials -Id $DeviceRequest.id

    }

    It -Skip "Approve a new device request" {
        $Response = PSc8y\Approve-DeviceRequest -Id $DeviceRequest.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceRequest -Id $DeviceRequest.id

    }
}

