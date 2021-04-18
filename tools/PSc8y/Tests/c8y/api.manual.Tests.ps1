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
    }
}
