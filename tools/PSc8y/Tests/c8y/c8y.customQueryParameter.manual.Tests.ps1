. $PSScriptRoot/../imports.ps1

Describe -Name "c8y customQueryParameter parameter" {
    It "adds custom query parmeters to an outgoing request" {
        $output = c8y inventory list --customQueryParam "myValue: 1" --dry --dryFormat json
        $request = $output | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $request.pathEncoded | Should -BeExactly "/inventory/managedObjects?myValue=1"
    }
}
