. $PSScriptRoot/../imports.ps1

Describe -Name "c8y data common parameter" {

    It "sets nested json objects via dot notation" {
        $output = c8y devices create --name test --data "my.nested.value=1" --dry --dryFormat json
        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            c8y_IsDevice = @{}
            my = @{
                nested = @{
                    value = 1
                }
            }
            name = "test"
        }
    }

    AfterEach {
    }
}
