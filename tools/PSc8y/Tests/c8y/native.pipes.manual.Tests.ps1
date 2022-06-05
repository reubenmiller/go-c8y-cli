. $PSScriptRoot/../imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    Context "pipeline - devices" {
        It "pipes objects from powershell to the c8y binary" {
            $output = 1..2 `
            | Invoke-ClientIterator "device" `
            | c8y devices create --template "{ jobIndex: input.index, name: input.value.name }" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path

            $requests[0].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; name="device0001"}
            $requests[1].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; name="device0002"}

            $requests[0].body.name | Should -BeOfType [string]
            $requests[0].body.jobIndex | Should -BeOfType [long]
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
