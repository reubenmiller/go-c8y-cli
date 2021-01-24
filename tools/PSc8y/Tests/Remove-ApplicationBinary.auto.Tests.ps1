. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ApplicationBinary" {
    BeforeEach {
        $app = New-TestHostedApplication
        New-ApplicationBinary -Id $app -File
        $appBinary = Get-ApplicationBinaryCollection -Id $App.id

    }

    It -Skip "Remove an application binary related to a Hosted (web) application" {
        $Response = PSc8y\Remove-ApplicationBinary -Application $app.id -BinaryId $appBinary.id
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Remove all application binaries (except for the active one) for an application (using pipeline)" {
        $Response = PSc8y\Get-ApplicationBinaryCollection -Id $app.id | Remove-ApplicationBinary -Application $app.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-Application -Id $app.id

    }
}

