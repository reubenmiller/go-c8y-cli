. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ChildAddition" {
    BeforeEach {
        $software = PSc8y\New-ManagedObject -Name softwarePackage1
        $version = PSc8y\New-ManagedObject -Name softwareVersion1
        PSc8y\Add-ChildAddition -Id $software.id -Child $version.id

    }

    It "Unassign a child addition from its parent managed object" {
        $Response = PSc8y\Remove-ChildAddition -Id $software.id -Child $version.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $version.id
        PSc8y\Remove-ManagedObject -Id $software.id

    }
}

