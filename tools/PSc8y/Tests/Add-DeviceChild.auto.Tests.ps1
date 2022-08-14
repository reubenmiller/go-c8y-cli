. $PSScriptRoot/imports.ps1

Describe -Name "Add-DeviceChild" {
    BeforeEach {

    }

    It "Add a related managed object as a child addition to an existing managed object" {
        $Response = PSc8y\Add-DeviceChild -Id $software.id -Child $version.id -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

