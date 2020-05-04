. $PSScriptRoot/imports.ps1

Describe -Name "New-Agent" {
    BeforeEach {
        $AgentName = PSc8y\New-RandomString -Prefix "myAgent"

    }

    It "Create agent" {
        $Response = PSc8y\New-Agent -Name $AgentName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create agent with custom properties" {
        $Response = PSc8y\New-Agent -Name $AgentName -Data @{ myValue = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Agent -Id $AgentName

    }
}

