. $PSScriptRoot/imports.ps1

Describe -Name "New-ExternalId" {
    BeforeEach {
        $my_SerialNumber = New-RandomString -Prefix "my_SerialNumber"
        $Device = New-TestDevice

    }

    It "Create external identity" {
        $Response = PSc8y\New-ExternalId -Device $Device.id -Type "$my_SerialNumber" -Name "myserialnumber"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create external identity (using pipeline)" {
        $Response = PSc8y\Get-Device $Device.id | New-ExternalId -Type "$my_SerialNumber" -Name "myserialnumber"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}

