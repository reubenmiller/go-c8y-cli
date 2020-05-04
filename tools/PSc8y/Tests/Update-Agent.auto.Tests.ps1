. $PSScriptRoot/imports.ps1

Describe -Name "Update-Agent" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It "Update agent by id" {
        $Response = PSc8y\Update-Agent -Id $agent.id -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update agent by name" {
        $Response = PSc8y\Update-Agent -Id $agent.name -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update agent custom properties" {
        $Response = PSc8y\Update-Agent -Id $agent.name -Data @{ "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $agent.id

    }
}

