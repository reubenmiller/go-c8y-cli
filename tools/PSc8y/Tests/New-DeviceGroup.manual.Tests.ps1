. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceGroup" {
    BeforeEach {
        $GroupName = PSc8y\New-RandomString -Prefix "mygroup"
    }

    It "Create device group" {
        $Response = PSc8y\New-DeviceGroup -Name $GroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.c8y_IsDeviceGroup | Should -Not -BeExactly $null
        $Response.type | Should -BeExactly "c8y_DeviceGroup"
    }

    It "Create device sub group" {
        $Response = PSc8y\New-DeviceGroup -Name $GroupName -Type c8y_DeviceSubGroup
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.c8y_IsDeviceGroup | Should -Not -BeExactly $null
        $Response.type | Should -BeExactly "c8y_DeviceSubGroup"
    }

    It "Create a device group with a custom type" {
        $Response = PSc8y\New-DeviceGroup -Name $GroupName -Data @{ type = "c8y_MyCustomGroup"; "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.type | Should -BeExactly "c8y_MyCustomGroup"
        $Response.myValue.value1 | Should -BeExactly $true
    }


    AfterEach {
        Remove-DeviceGroup -Id $GroupName

    }
}

