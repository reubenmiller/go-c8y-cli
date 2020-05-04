. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceGroup" {
    BeforeEach {
        $GroupName = PSc8y\New-RandomString -Prefix "mygroup"

    }

    It "Create device group" {
        $Response = PSc8y\New-DeviceGroup -Name $GroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create device group with custom properties" {
        $Response = PSc8y\New-DeviceGroup -Name $GroupName -Data @{ "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-DeviceGroup -Id $GroupName

    }
}

