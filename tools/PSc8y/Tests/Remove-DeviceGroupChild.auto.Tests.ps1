. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceGroupChild" {
    BeforeEach {

    }

    It -Skip "Unassign a child addition from its parent managed object" {
        $Response = PSc8y\Remove-DeviceGroupChild -Id $software.id -Child $version.id -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

