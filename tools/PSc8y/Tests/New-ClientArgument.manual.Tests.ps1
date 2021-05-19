. $PSScriptRoot/imports.ps1

InModuleScope PSc8y {
    Describe "New-ClientArgument" {
        It "Accepts an array of values" {
            $c8yargs = New-ClientArgument -Parameters @{"Ids" = @("1111", "2222")}
            $c8yargs | Should -BeExactly @("--ids=`"1111,2222`"", "--allowEmptyPipe")
        }

        It "handles an empty array" {
            $c8yargs = New-ClientArgument -Parameters @{"Ids" = @()}
            $c8yargs | Should -BeExactly @("--allowEmptyPipe")
        }

        It "handles an array with a single value" {
            $c8yargs = New-ClientArgument -Parameters @{"Ids" = @("1111")}
            $c8yargs | Should -BeExactly @("--ids=1111", "--allowEmptyPipe")
        }

        It "handles an array of objects picking out the id" {
            $c8yargs = New-ClientArgument -Parameters @{"id" = @(
                @{id="1111"},
                @{id="2222"}
            )}
            $c8yargs | Should -BeExactly @("--id=`"1111,2222`"", "--allowEmptyPipe")
        }

        It "Converts hashtables to escapped json" {
            $Parameters = @{
                complex = @{"id" = 1}
            }
            $c8yargs = New-ClientArgument -Parameters:$Parameters
            $c8yargs | Should -HaveCount 3
            $c8yargs[0] | Should -BeExactly '--complex'
            $c8yargs[1] | Should -BeExactly '{\"id\":1}'
            $c8yargs[2] | Should -BeExactly "--allowEmptyPipe"
        }
    }
}
