. $PSScriptRoot/imports.ps1

Describe -Name "Get-ManagedObjectChild" {
    BeforeEach {

    }

    It -Skip "Get an existing child managed object" {
        $Response = PSc8y\Get-ManagedObjectChild -Id $Agent.id -Child $Ref.id -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

