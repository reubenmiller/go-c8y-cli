. $PSScriptRoot/imports.ps1

Describe -Name "Disable-Application" {
    BeforeEach {

    }

    It -Skip "Disable an application of a tenant" {
        $Response = PSc8y\Disable-Application -Tenant t12345 -Application myMicroservice
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

