. $PSScriptRoot/imports.ps1

Describe -Name "Update-User" {
    Context "existing user" {
        BeforeAll {
            $User = PSc8y\New-TestUser
        }

        It "Update custom properties for a user" {
            $Response = PSc8y\Update-User -Id $User.id -CustomProperties @{
                language = "fr"
            }
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.customProperties.language | Should -BeExactly "fr"
        }

        It "Update user (from pipeline)" {
            $Response = $User | PSc8y\Update-User -CustomProperties @{
                language = "de"
            }
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.customProperties.language | Should -BeExactly "de"
        }

        It "Shows the request to be sent when disabling an existing user sends a boolean" {
            $output = PSc8y\Update-User -Id $User.id -Enabled:$false -WhatIf 2>&1
            $LASTEXITCODE | Should -Be 0

            $Body = Get-RequestBodyCollection $output
            $Body | Should -MatchObject @{ enabled = $false }
        }

        It "Disable an existing user" {
            $Response = PSc8y\Update-User -Id $User.id -Enabled:$false
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.enabled | Should -BeExactly $false
            $Response.enabled | Should -HaveType [boolean]
        }

        It "Enable an existing user" {
            $Response = PSc8y\Update-User -Id $User.id -Enabled
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.enabled | Should -BeExactly $true
        }

        # Note: passing an object to a string, does some weird type coersion.
        # Might have to switch to using an [object] as the type
        It -Skip "Update user (using object)" {
            $Response = PSc8y\Update-User -Id $User -CustomProperties @{
                language = "de"
            }
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.customProperties.language | Should -BeExactly "de"
        }

        AfterAll {
            Remove-User -Id $User.id
        }
    }
}
