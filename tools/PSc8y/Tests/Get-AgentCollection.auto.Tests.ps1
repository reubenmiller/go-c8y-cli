. $PSScriptRoot/imports.ps1

Describe -Name "Get-AgentCollection" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It -Skip "Get a collection of agents with type 'myType', and their names start with 'sensor'" {
        $Response = PSc8y\Get-AgentCollection -Name "sensor*" -Type myType
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $agent.id

    }
}

