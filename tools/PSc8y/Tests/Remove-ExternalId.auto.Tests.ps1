. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ExternalId" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ExternalID = PSc8y\New-ExternalId -Device $Device.id -Type "my_SerialNumber" -Name "myserialnumber2"

    }

    It "Delete external identity" {
        $Response = PSc8y\Remove-ExternalId -Type "my_SerialNumber" -Name "myserialnumber2"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

