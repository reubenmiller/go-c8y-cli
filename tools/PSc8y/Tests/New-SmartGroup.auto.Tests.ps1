. $PSScriptRoot/imports.ps1

Describe -Name "New-SmartGroup" {
    BeforeEach {
        $smartgroupName = PSc8y\New-RandomString -Prefix "mySmartGroup_createTests"

    }

    It "Create smart group (without a filter)" {
        $Response = PSc8y\New-SmartGroup -Name $smartgroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create smart group with a device filter (filter by type and has a serial number)" {
        $Response = PSc8y\New-SmartGroup -Name $smartgroupName -Query "type eq 'IS*' and has(c8y_Hardware.serialNumber)"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create a smart group which is not visible in the UI" {
        $Response = PSc8y\New-SmartGroup -Name $smartgroupName -Query "type eq 'IS*'" -Invisible
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-SmartGroup -Id $smartgroupName

    }
}

