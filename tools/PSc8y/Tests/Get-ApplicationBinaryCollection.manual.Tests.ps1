#. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationBinaryCollection" {
    Context "existing web application" {
        BeforeAll {
            $application = New-TestHostedApplication
        }

        It -Skip "Gets a list of binaries for a given application" {
            $application.id | Should -Not -BeNullOrEmpty
            [array] $response = Get-ApplicationBinaryCollection -Id $application.id

            $LASTEXITCODE | Should -Be 0
            $application | Should -Not -BeNullOrEmpty
            $response | Should -HaveCount 1

            $application.activeVersionId | Should -BeExactly $response[0].id
        }

        AfterAll {
            if ($application) {
                PSc8y\Remove-Application -Id $application.id
            }
        }
    }
}
