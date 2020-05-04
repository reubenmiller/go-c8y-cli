. $PSScriptRoot/imports.ps1

Describe -Name "Update-Application" {
    BeforeEach {
        $App = New-Application -Name "helloworld-app" -Type HOSTED -Key "helloworld-app-key" -ContextPath "helloworld-app"

    }

    It "Update application availability to MARKET" {
        $Response = PSc8y\Update-Application -Id "helloworld-app" -Availability "MARKET"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $App.id

    }
}

