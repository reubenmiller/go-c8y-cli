. $PSScriptRoot/imports.ps1

Describe -Name "New-Device" {
    BeforeEach {
        $DeviceName = PSc8y\New-RandomString -Prefix "myDevice"

    }

    It "Create device" {
        $Response = PSc8y\New-Device -Name $DeviceName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create device with custom properties" {
        $Response = PSc8y\New-Device -Name $DeviceName -Data @{ myValue = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create device using a template" {
        $Response = PSc8y\New-Device -Template "{ name: '$DeviceName' }"

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Device -Id $DeviceName

    }
}

