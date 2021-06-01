. $PSScriptRoot/../imports.ps1

Describe -Name "c8y api" {

    Context "Custom POST requests" {
        It "Allows non-json bodies" {
            $output = c8y api POST /myvalue --data "myvalue,41,outputtext" --contentType "text/plain" --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $request = $output | ConvertFrom-Json
            $request | Should -MatchObject @{
                method = "POST"
                path = "/myvalue"
                body = "myvalue,41,outputtext"
            } -Property body, method, path
        }

        It "Allows shorthand json bodies" {
            $output = c8y api POST /myvalue --data "myvalue=1" --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $request = $output | ConvertFrom-Json
            $request | Should -MatchObject @{
                method = "POST"
                path = "/myvalue"
                body = @{
                    myvalue = 1
                }
            } -Property body, method, path
        }

        It "accepts paths via pipeline" {
            $paths = @(
                "/inventory/managedObjects?pageSize=1&withTotalPages=true",
                "/application/applications?pageSize=2"
            )
            $output = $paths | c8y api --dry --dryFormat json

            $requests = $output | ConvertFrom-Json
            $LASTEXITCODE | Should -Be 0
            $requests[0].path | Should -BeExactly "/inventory/managedObjects"
            $requests[0].query | Should -BeExactly "pageSize=1&withTotalPages=true"
            $requests[1].path | Should -BeExactly "/application/applications"
            $requests[1].query | Should -BeExactly "pageSize=2"
        }

        It "adds custom query parmeters to an outgoing request" {
            $output = "/inventory/managedObjects?pageSize=1" | c8y api --customQueryParam "myValue=2" --dry --dryFormat json
            $request = $output | ConvertFrom-Json
            $LASTEXITCODE | Should -Be 0
            $queryParts = $request.query -split "&" | Sort-Object
            $queryParts[0] | Should -BeExactly "myValue=2"
            $queryParts[1] | Should -BeExactly "pageSize=1"
            $request.path | Should -BeLikeExactly "/inventory/managedObjects"
            # $request.pathEncoded | Should -BeExactly "/inventory/managedObjects?pageSize=1&myValue=1"
        }

        It "accepts positional arguments for method and path (not using pipeline)" {
            $output = c8y api GET "/alarm/alarms?pageSize=10&status=ACTIVE" --dry --dryFormat json `
            | ConvertFrom-Json
            $output.method | Should -BeExactly "GET"
            $output.path | Should -BeExactly "/alarm/alarms"
            $output.query | Should -BeExactly "pageSize=10&status=ACTIVE"
        }

        It "accepts positional arguments for path and defaults to GET (not using pipeline)" {
            $output = c8y api "/alarm/alarms" --dry --dryFormat json `
            | ConvertFrom-Json
            $output.method | Should -BeExactly "GET"
            $output.path | Should -BeExactly "/alarm/alarms"
            $output.query | Should -BeNullOrEmpty
        }

        It "accepts positional path argument and explicit method" {
            $output = c8y api "/alarm/alarms" --method post --dry --dryFormat json `
            | ConvertFrom-Json
            $output.method | Should -BeExactly "POST"
            $output.path | Should -BeExactly "/alarm/alarms"
            $output.query | Should -BeNullOrEmpty
        }
    }
}
