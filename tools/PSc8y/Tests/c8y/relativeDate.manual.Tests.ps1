. $PSScriptRoot/../imports.ps1

Describe -Name "c8y relative datetime" {

    Context "alarms" {
        It "converts a relative date to a iso8601 formatted string" {

            $output = c8y alarms list --dateFrom '-1d' --dry
            @($output -match 'dateFrom').Count | Should -BeGreaterThan 0
        }
    }

    AfterEach {
    }
}
