. $PSScriptRoot/imports.ps1

Describe -Name "New-Microservice" {
    Context "Test Microservice" {
        BeforeAll {
            $AppName = New-RandomString -Prefix "testms-"
            $MicroserviceZip = "$PSScriptRoot/TestData/microservice/helloworld.zip"
        }

        It "Creates a new microservice using the name from the zip file" {
            # Create copy of example microservice zip file
            $Name = New-RandomString -Prefix "testms-"
            $CustomZip = Copy-Item $MicroserviceZip -Destination "${Name}.zip" -PassThru

            # Remove microservice (if exists)
            Get-Microservice -Id $Name | Remove-Microservice

            $App = New-Microservice -File $CustomZip

            # Remove temp file
            if ($CustomZip) {
                Remove-Item $CustomZip -Force
            }

            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.name | Should -BeExactly $Name
        }

        It "Creates a new microservice from a zip file with a custom name" {
            # Note: Cumulocity trial tenant does not support microservice hosting, so the binary can't be uploaded withouth a 403 error
            $App = New-Microservice -Name $AppName -File $MicroserviceZip
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.name | Should -BeExactly $AppName
        }

        It "Update existing (enabled) microservice" {
            $AppBeforeUpdate = Get-Application -Id $AppName
            $AppBeforeUpdate | Should -Not -BeNullOrEmpty

            $App = New-Microservice -Name $AppName -File $MicroserviceZip
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.id | Should -BeExactly $AppBeforeUpdate.id
        }

        It "Create microservice placeholder but skipping upload" {
            Get-Microservice -Id $AppName | Remove-Microservice

            $ManifestFile = New-TemporaryFile
            $ManifestFile = $ManifestFile | Rename-Item -NewName { $_.name + ".json" } -PassThru

            Out-File -FilePath $ManifestFile -InputObject @"
{
    "apiVersion": "v1",
    "name": "helloworld",
    "version": "1.0.0",
    "provider": {
        "name": "New Company Ltd.",
        "domain": "http://new-company.com",
        "support": "support@new-company.com"
    },
    "isolation": "PER_TENANT",
    "requiredRoles": [],
    "livenessProbe": {
        "httpGet": {
            "path": "/health"
        },
        "initialDelaySeconds": 60,
        "periodSeconds": 10
    },
    "readinessProbe": {
        "httpGet": {
            "path": "/health",
            "port": 80

        },
        "initialDelaySeconds": 20,
        "periodSeconds": 10
    }
}
"@

            $App = New-Microservice -Name $AppName -File $ManifestFile -SkipUpload
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.id | Should -MatchExactly '^\d+$'

            # Check credentials
            $BootstrapUser = Get-MicroserviceBootstrapUser -Id $App.id
            $BootstrapUser.tenant | Should -BeExactly $env:C8Y_TENANT
            $BootstrapUser.name | Should -BeLike "service*"
            $BootstrapUser.password | Should -Not -BeNullOrEmpty

            # Check manifest
            $App = Get-Microservice -Id $AppName
            $App.manifest.requiredRoles | Should -BeExactly @()
        }

        It "Trying creating microservice with invalid manifest json" {
            Get-Microservice -Id $AppName | Remove-Microservice

            $ManifestFile = New-TemporaryFile

            Out-File -FilePath $ManifestFile -InputObject @"
Invalid json example
"@

            $App = New-Microservice -Name $AppName -File $ManifestFile -SkipUpload -ErrorVariable ErrorResponse
            $LASTEXITCODE | Should -Not -Be 0
            $ErrorResponse | Select-Object -Last 1 | Should -BeLike "*invalid manifest*"
            $App | Should -BeNullOrEmpty
        }

        It "Creates a new microservice but does not subscribe to it automatically" {
            Get-Microservice -Id $AppName | Remove-Microservice

            $App = New-Microservice -Name $AppName -File $MicroserviceZip -SkipSubscription
            $LASTEXITCODE | Should -Be 0
            $App | Should -Not -BeNullOrEmpty
            $App.name | Should -BeExactly $AppName

            # TODO: Check if the application is subscribe to or not
            Enable-Application -Application $AppName
            $LASTEXITCODE | Should -Be 0
        }

        AfterAll {
            Remove-Microservice -Id $AppName -ErrorAction SilentlyContinue
        }
    }
}
