. $PSScriptRoot/imports.ps1

Describe -Name "Get-AgentCollection" {
    Context "Agents with spaces in their names" {
        BeforeAll {
            $RandomPart = New-RandomString
            $Agent01 = New-TestAgent -Name "My Custom Agent $RandomPart"
            $Agent02 = New-TestAgent -Name "My Custom Agent $RandomPart"
        }

        It "Find devices by name" {
            $Response = PSc8y\Get-AgentCollection -Name "*My Custom Agent ${RandomPart}*" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.Count | Should -BeExactly 2
            @($Response.id -match $Agent01.id) | Should -HaveCount 1
            @($Response.id -match $Agent02.id) | Should -HaveCount 1
        }

        AfterAll {
            $null = Remove-ManagedObject -Id $Agent01.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-ManagedObject -Id $Agent02.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
