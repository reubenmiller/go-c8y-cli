. $PSScriptRoot/imports.ps1

Describe -Name "New-Firmware" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a firmware package" {
        $Response = PSc8y\New-Firmware -Name "iot-linux" -Description "Linux image for IoT devices" -Data @{$type=@{}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -FragmentType $type | Remove-ManagedObject

    }
}

