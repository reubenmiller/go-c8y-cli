. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Agent" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It -Skip "Remove agent by id" {
        $Response = PSc8y\Remove-Agent -Id $agent.id
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Remove agent by name" {
        $Response = PSc8y\Remove-Agent -Id $agent.name
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete agent and related device user/credentials" {
        $Response = PSc8y\Remove-Agent -Id "agent01" -WithDeviceUser
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

