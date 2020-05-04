. $PSScriptRoot/imports.ps1

Describe -Name "Update-Agent" {
    BeforeEach {
        $agent = PSc8y\New-TestAgent

    }

    It "Update device by id" {
        $Response = PSc8y\Update-Agent -Id $agent.id -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.name | Should -BeExactly "MyNewName"
    }

    AfterEach {
        Remove-ManagedObject -Id $agent.id

    }
}

