. $PSScriptRoot/imports.ps1

Describe -Name "New-ApplicationVersion" {
    BeforeEach {

    }

    It -Skip "Create a new application version" {
        $Response = PSc8y\New-ApplicationVersion -Name $AppName -Key "${AppName}-key" -ContextPath $AppName -Type HOSTED
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

