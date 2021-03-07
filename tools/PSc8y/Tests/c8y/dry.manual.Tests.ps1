. $PSScriptRoot/../imports.ps1

Describe -Name "c8y dry" {

    Context "multi-form data" {
        It "Displays multi form data" {
            $tempFile = New-TemporaryFile
            "äüText" | Out-File $tempFile
            $output = c8y binaries create --file=$tempFile --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }
    }

    Context "POST" {
        It "Shows the body contents" {

            $output = c8y devices create --name test01 --dry --withError
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }

        It "Shows the body contents with a custom body" {

            $output = c8y devices create --name test01 --data "test=1" --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
        }

        It "Hides sensitive information" {
            $env:C8Y_LOGGER_HIDE_SENSITIVE = "true"
            $output = c8y devices create --name test01 --data "test=1" --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
            $output | Should -Not -Match $env:C8Y_USERNAME

            $env:C8Y_LOGGER_HIDE_SENSITIVE = $null
        }

        
    }

    AfterEach {
    }
}
