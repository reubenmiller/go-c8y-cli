. $PSScriptRoot/imports.ps1

Describe -Name "New-Microservice" {
    Context "Test Microservice" {
        $AppName = New-RandomString -Prefix "testms-"
        $MicroserviceZip = "$PSScriptRoot/TestData/microservice/helloworld.zip"

        <# Notes:
            * Cumulocity trial tenant does not support microservice hosting, so the binary can't be uploaded withouth a 403 error
        #>

        It "Creates a new microservice from a zip file" {
            # Note: Cumulocity trial tenant does not support microservice hosting, so the binary can't be uploaded withouth a 403 error
            $App = New-Microservice -Name $AppName -File $MicroserviceZip -SkipUpload
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.name | Should -BeExactly $AppName
        }

        It "Update existing (enabled) microservice" {
            $AppBeforeUpdate = Get-Application -Id $AppName
            $AppBeforeUpdate | Should -Not -BeNullOrEmpty

            $App = New-Microservice -Name $AppName -File $MicroserviceZip -SkipUpload
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.id | Should -BeExactly $AppBeforeUpdate.id
        }

        <# It "create new microservice with name from zip file" {
            throw "not implemented"
        } #>

        Remove-Microservice -Id $AppName
    }
}
