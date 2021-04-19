. $PSScriptRoot/../imports.ps1

Describe -Name "c8y relative datetime" {

    Context "alarms" {
        It "converts a relative date to a iso8601 formatted string" {

            $output = c8y alarms list --dateFrom '-1d' --dry
            @($output -match 'dateFrom').Count | Should -BeGreaterThan 0
        }
    }

    Context "events" {
        It "does not encode the plus sign if used in the body" {
            $output = c8y events create --device 1234 --type "myType" --text "example" --time "2021-04-19T22:57:38.41129191+02:00" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0
            $request = $output | ConvertFrom-Json
            $request.body.time | Should -BeExactly "2021-04-19T22:57:38.41129191+02:00"
        }

        It "encodes the plus sign if used in query parameters" {
            $output = c8y events list --dateFrom "2021-04-19T22:57:38.41129191+02:00" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0
            $request = $output | ConvertFrom-Json
            $request.pathEncoded | Should -BeExactly "/event/events?dateFrom=2021-04-19T22:57:38.41129191%2B02:00"
        }

        It "does not encode the plus sign when using inside a jsonnet template" {
            $output = c8y events create --device 1234 --type "myType" --text "example" --time "2021-04-19T22:57:38.41129191+02:00" --template "{mtime: _.NowNano('2022-04-19T22:57:38.41129191+02:00')}" --dry --dryFormat json
            
            $LASTEXITCODE | Should -BeExactly 0
            $request = $output | ConvertFrom-Json
            $request.body.time | Should -BeExactly "2021-04-19T22:57:38.41129191+02:00"
            $request.body.mtime | Should -BeExactly "2022-04-19T22:57:38.41129191+02:00"
        }
    }

    AfterEach {
    }
}
