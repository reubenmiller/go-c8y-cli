. $PSScriptRoot/imports.ps1

Describe -Name "Disable-Application" {
    BeforeEach {

    }

    It -Skip "Disable an application of a tenant" {
        $Response = PSc8y\Disable-Application -Tenant mycompany -Application myMicroservice
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

