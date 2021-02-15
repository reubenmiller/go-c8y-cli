. $PSScriptRoot/../imports.ps1

Describe -Name "json" {
    Context "color" {
        It "Does not colorize json when being assigned to a variable" -TestCases @(
            @{option = ""},
            @{option = "--noColor"},
            @{option = "--noColor=true"}
        ) {
            Param([string]$option)

            $output = c8y applications list --pageSize 1 $option
            $output | should -Not -Match "\x1b\[[0-9;]*m"
        }

        It "prints json in color" {
            $output = c8y applications list --pageSize 1 --noColor=false
            $output | should -Match "\x1b\[[0-9;]*m"
        }

        It "should not print json in color when using pretty print and no color" {
            $output = c8y applications list --noColor=true --compact=false
            $output | Out-String | Should -Not -Match "\x1b\[[0-9;]*m"
        }

        It "should not print json in color when using pretty print and no color and no streaming" {
            $output = c8y applications list --noColor=true --stream=false --compact=false
            $output | Out-String | Should -Not -Match "\x1b\[[0-9;]*m"
        }

        It "does not print in color when csv is being used" {
            $output = c8y applications get --id cockpit --select id,name --csv --compact=false
            $output | Out-String | Should -Not -Match "\x1b\[[0-9;]*m"
        }
    }

    Context "compact" {
        It "Prints compact json when being assigned to a variable (<option>)" -TestCases @(
            @{option = ""},
            @{option = "--compact"},
            @{option = "--compact=true"}
        ) {
            Param([string]$option)

            $output = c8y applications list --pageSize 1 $option
            $output | Should -BeOfType string
            $output -split "\n" | Should -HaveCount 1
            $output | Should -BeLike "{*}"
        }

        It "Print pretty json when not using compact" {
            $output = c8y applications list --pageSize 1 --compact=false
            ($output -split "\n").Count | Should -BeGreaterThan 5
            ($output | Out-String).Trim() | Should -BeLike "{*}"
        }

        It "Print the raw json output in compact json" {
            $output = c8y applications list --pageSize 1 --compact --raw
            $output -split "\n" | Should -HaveCount 1
            $output | Should -BeLike "{*}"
            $json = $output | ConvertFrom-Json
            $json.applications | Should -HaveCount 1
        }

        It "Only returns properties for given fields and prints them in pretty print" {
            $output = c8y applications list --select id,name --pageSize 2 --compact=false
            $output | Should -HaveCount 8

            # Convert json lines to array of json text
            $jsonlines = ($output | out-string) -split "(?<=\n})\n"
            $json = $jsonlines | ConvertFrom-Json
            $json.id | Should -HaveCount 2
        }

        It "Only returns properties for given fields and prints compact json" {
            $output = c8y applications list --select id,name --pageSize 2
            $json = $output | ConvertFrom-Json
            $output | Should -HaveCount 2
            $json.id | Should -HaveCount 2
        }

        It "Prints the raw json in pretty print" {
            $output = c8y applications list --pageSize 1 --raw --compact=false
            $output.Count | Should -BeGreaterThan 5
            ($output | Out-String).Trim() | Should -BeLike "{*}"
            $output | ConvertFrom-Json | Should -Not -BeNullOrEmpty
        }

        It "Ignores settings when csv is being used" {
            $output = c8y applications get --id cockpit --select id,name --csv --compact=false
            $output | Should -MatchExactly "^\d+,cockpit$"
        }
    }

    Context "stream" {
        It "Prints json lines when using stream parameter (<option>)" -TestCases @(
            @{option = ""},
            @{option = "--stream"},
            @{option = "--stream=true"}
        ) {
            param([string] $option)
            $output = c8y applications list --pageSize 2 $option
            $output | Should -BeOfType string
            $output | Should -HaveCount 2
            $output | ForEach-Object {
                $_ | Should -BeLike "{*}"
            }
            $json = $output | ConvertFrom-Json
            $json | Should -Not -BeNullOrEmpty
            $json | Should -HaveCount 2
        }

        It "Prints the json array (not json lines)" {
            $output = c8y applications list --pageSize 1 --stream=false --compact=false
            ($output -split "\n").Count | Should -BeGreaterThan 5
            ($output | Out-String).Trim() | Should -BeLike '`[*`]'
            $output | ConvertFrom-Json | Should -Not -BeNullOrEmpty
        }
    }
}
