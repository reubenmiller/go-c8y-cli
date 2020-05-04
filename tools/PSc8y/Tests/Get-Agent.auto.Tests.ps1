. $PSScriptRoot/imports.ps1

Describe -Name "Get-Agent" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It "Get agent by id" {
        $Response = PSc8y\Get-Agent -Id $agent.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get agent by name" {
        $Response = PSc8y\Get-Agent -Id $agent.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $agent.id

    }
}

