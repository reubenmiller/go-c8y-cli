. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ExternalId" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ExternalID = PSc8y\New-ExternalId -Device $Device.id -Type "my_SerialNumber" -Name "myserialnumber2"

    }

    It -Skip "Delete external identity" {
        $Response = PSc8y\Remove-ExternalId -Type "my_SerialNumber" -Name "myserialnumber2"
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete a specific external identity type (via pipeline)" {
        $Response = PSc8y\Get-DeviceCollection | Get-ExternalIdCollection -Filter 'type eq c8y_Serial' | Remove-ExternalId -Type c8y_Serial
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

