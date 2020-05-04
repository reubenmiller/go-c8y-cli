. $PSScriptRoot/imports.ps1

Describe -Name "Get-ExternalIdCollection" {
    BeforeEach {
        $Device = New-TestDevice
        $ExtName = New-RandomString -Prefix "IMEI"
        $ExternalID = PSc8y\New-ExternalId -Device $Device.id -Type "my_SerialNumber" -Name "$ExtName"

    }

    It "Get a list of external ids" {
        $Response = PSc8y\Get-ExternalIdCollection -Device $Device.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}

