. $PSScriptRoot/imports.ps1

Describe -Name "Test-Json" {

    It "should detect invalid json" {
        $Response = Test-Json "{ value = 1 }" -InformationVariable Messages
        $Response | Should -BeExactly $false
        $Messages | Should -Not -BeNullOrEmpty
        $Messages | Should -BeLike "Invalid json*"
    }

    It "should detect invalid json arrays" {
        $Response = Test-Json " [ value = 1 ]  " -InformationVariable Messages
        $Response | Should -BeExactly $false
        $Messages | Should -Not -BeNullOrEmpty
        $Messages | Should -BeLike "Invalid json*"
    }

    It "should detect valid json" {
        $Response = Test-Json '{ "value": "1" }' -InformationVariable Messages
        $Response | Should -BeExactly $true
        $Messages | Should -BeNullOrEmpty
    }

    It "should not allow json literals (i.e. non json objects/arrays)" {
        $Response = Test-Json "true" -InformationVariable Messages
        $Response | Should -BeExactly $false
        $Messages | Should -BeLikeExactly "Only json array or objects are supported"
    }
}
