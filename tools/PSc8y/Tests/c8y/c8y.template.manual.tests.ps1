. $PSScriptRoot/../imports.ps1

Describe -Name "c8y template" {
    Context "Template" {
        It "template should preservce double quotes" {
            $output = c8y template execute --template '{\"email\": \"he ll@ex ample.com\"}'
            $LASTEXITCODE | Should -Be 0
            $body = $output | ConvertFrom-Json
            $body.email | Should -MatchExactly "he ll@ex ample.com"
        }

    }
}
