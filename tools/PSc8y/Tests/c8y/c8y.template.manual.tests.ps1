. $PSScriptRoot/../imports.ps1

Describe -Name "c8y template" {
    It "template should preservce double quotes" {
        $output = c8y template execute --template '{\"email\": \"he ll@ex ample.com\"}'
        $LASTEXITCODE | Should -Be 0
        $body = $output | ConvertFrom-Json
        $body.email | Should -MatchExactly "he ll@ex ample.com"
    }

    It "provides relative time functions" {
        $output = c8y template execute --template "{now: _.Now(), nowRelative: _.Now('-1h'), nowNano: _.NowNano(), nowNanoRelative: _.NowNano('-10d')}"
        $LASTEXITCODE | Should -Be 0
        $data = $output | ConvertFrom-Json
        Get-Date $data.now | Should -Not -BeNullOrEmpty
        Get-Date $data.nowRelative | Should -Not -BeNullOrEmpty
        Get-Date $data.nowNano | Should -Not -BeNullOrEmpty
        Get-Date $data.nowNanoRelative | Should -Not -BeNullOrEmpty
    }
}
