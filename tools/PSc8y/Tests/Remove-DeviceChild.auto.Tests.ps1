. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceChild" {
    BeforeEach {

    }

    It -Skip "Unassign a child addition from its parent managed object" {
        $Response = PSc8y\Remove-DeviceChild -Id $software.id -Child $version.id -ChildType addition
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

