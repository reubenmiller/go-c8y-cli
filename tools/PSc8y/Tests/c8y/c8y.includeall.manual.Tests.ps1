. $PSScriptRoot/../imports.ps1

Describe -Name "c8y select global parameter" {
    BeforeEach {
        $prefix = New-RandomString
        $type = New-RandomString
        $mos = c8y inventory create --type "$type" --name $prefix
    }

    It "does not return duplicate results when using includeAll" {
        $output = c8y inventory list --includeAll --type "$type" --filter "name like $prefix"
        $LASTEXITCODE | Should -Be 0
        $results = $output | ConvertFrom-Json
        $results | Should -HaveCount 1
    }

    AfterEach {
        $mos | c8y inventory delete
    }
}
