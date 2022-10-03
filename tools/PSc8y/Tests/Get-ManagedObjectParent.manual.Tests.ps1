. $PSScriptRoot/imports.ps1

Describe -Name "Get-ManagedObjectParent" {
    BeforeEach {

    }

    It -Skip "Get addition parent" {
        $Response = PSc8y\Get-AdditionParent -Id $mo.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

