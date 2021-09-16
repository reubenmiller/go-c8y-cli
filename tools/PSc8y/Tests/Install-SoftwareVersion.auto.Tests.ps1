. $PSScriptRoot/imports.ps1

Describe -Name "Install-SoftwareVersion" {
    BeforeEach {

    }

    It -Skip "Get a software package" {
        $Response = PSc8y\Install-SoftwareVersion -Device $mo.id -Software go-c8y-cli -Version 1.0.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

