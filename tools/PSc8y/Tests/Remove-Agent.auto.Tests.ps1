. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Agent" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It "Remove agent by id" {
        $Response = PSc8y\Remove-Agent -Id $agent.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove agent by name" {
        $Response = PSc8y\Remove-Agent -Id $agent.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

