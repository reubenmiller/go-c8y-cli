. $PSScriptRoot/../imports.ps1

Describe -Name "c8y dry" {

    Context "multi-form data" {
        It "Displays multi form data" {
            $tempFile = New-TemporaryFile
            "äüText" | Out-File $tempFile
            $output = c8y binaries create --file=$tempFile --dry
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }
    }

    Context "POST" {
        It "Shows the body contents" {

            $output = c8y devices create --name test01 --dry
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }

        It "Shows the body contents with a custom body" {

            $output = c8y devices create --name test01 --data "test=1" --dry
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }

        It "Hides sensitive information" {
            $env:C8Y_SETTINGS_LOGGER_HIDESENSITIVE = "true"
            $output = c8y devices create --name test01 --data "test=1" --dry
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
            $output | Should -Not -Match [regex]::Escape($env:C8Y_USERNAME)

            $env:C8Y_SETTINGS_LOGGER_HIDESENSITIVE = $null
        }

        
    }

    AfterEach {
    }
}
