. $PSScriptRoot/imports.ps1

Describe -Name "New-Application" {
    BeforeEach {
        $AppName = New-RandomString -Prefix "testapp_"

    }

    It "Create a new hosted application" {
        $Response = PSc8y\New-Application -Name $AppName -Key "${AppName}-key" -ContextPath $AppName -Type HOSTED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $AppName

    }
}

