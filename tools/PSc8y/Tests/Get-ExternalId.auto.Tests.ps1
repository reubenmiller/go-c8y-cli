. $PSScriptRoot/imports.ps1

Describe -Name "Get-ExternalId" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $ExternalID = PSc8y\New-ExternalId -Device $Device.id -Type "my_SerialNumber" -Name "myserialnumber"

    }

    It "Get external identity" {
        $Response = PSc8y\Get-ExternalId -Type "my_SerialNumber" -Name "myserialnumber"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

