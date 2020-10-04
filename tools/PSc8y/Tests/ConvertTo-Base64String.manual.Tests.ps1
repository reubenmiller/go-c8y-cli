. $PSScriptRoot/imports.ps1

Describe -Name "ConvertTo-Base64String" {

    It "Convert to base64 encoded string and back" {

        $originalText = "t12345/test:muyPass word *!2'"
        $base64Text = $originalText | ConvertTo-Base64String

        $converted = $base64Text | ConvertFrom-Base64String

        $converted | Should -BeExactly $originalText
        $converted | Should -Not -BeExactly $base64Text
    }
}
