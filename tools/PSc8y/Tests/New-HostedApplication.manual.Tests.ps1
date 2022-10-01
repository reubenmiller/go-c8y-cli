. $PSScriptRoot/imports.ps1

Describe -Name "New-HostedApplication" {
    Context "developer environment 1" {
        BeforeEach {
            $WebAppSource = "$PSScriptRoot/TestData/hosted-application/simple-helloworld"
            $AppName = New-RandomString -Prefix "MyApp"
            $VerboseFile = New-TemporaryFile
        }

        It "Create web application (without uploading binary) and use custom settings" {
            $ContextPath = ($AppName -replace " ", "").ToLower()

            $application = PSc8y\New-HostedApplication `
                -Name $AppName `
                -ResourcesUrl "/subPath" `
                -Availability "MARKET" `
                -ContextPath $ContextPath `
                -Verbose 2> $VerboseFile

            $LASTEXITCODE | Should -Be 0
            $application | Should -Not -BeNullOrEmpty

            $application.name | Should -BeExactly $AppName
            $application.contextPath | Should -BeExactly $ContextPath
            $application.resourcesUrl | Should -BeExactly "/subPath"
            $application.activeVersionId | Should -BeNullOrEmpty

            $BodyInVerbose = Get-JSONFromResponse (Get-Content $VerboseFile -Raw)
            $BodyInVerbose.availability | Should -BeExactly "MARKET"

            $Binaries = Get-ApplicationBinaryCollection -Id $application.id -PageSize 100
            $Binaries | Should -BeNullOrEmpty
        }

        It "Create web application using dry run" {
            $ContextPath = ($AppName -replace " ", "").ToLower()

            $output = PSc8y\New-HostedApplication `
                -Name $AppName `
                -ResourcesUrl "/subPath" `
                -Availability "MARKET" `
                -ContextPath $ContextPath `
                -Dry -DryFormat json

            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
            $request = $output | ConvertFrom-Json
            $request.path | Should -BeExactly "/application/applications"
            $request.body.name | Should -BeExactly $AppName 
            $request.body.type | Should -BeExactly "HOSTED" 
            $request.body.resourcesUrl | Should -BeExactly "/subPath" 
            $request.body.availability | Should -BeExactly "MARKET" 

        }

        AfterEach {
            PSc8y\Get-Application -Id $AppName | PSc8y\Remove-Application

            Remove-Item $VerboseFile
        }

    }

    Context "developer environment 2" {
        BeforeEach {
            $WebAppSource = "$PSScriptRoot/TestData/hosted-application/simple-helloworld"
            $AppName = New-RandomString -Prefix "app"
        }

        It "Create a new web application from a folder" {
            $application = PSc8y\New-HostedApplication -File $WebAppSource -Name $AppName
            $LASTEXITCODE | Should -Be 0
            $application | Should -Not -BeNullOrEmpty
            $application.name | Should -BeExactly $AppName
            $application.contextPath | Should -BeExactly $AppName
            $application.resourcesUrl | Should -BeExactly "/"
            $application.activeVersionId | Should -MatchExactly "^\d+$"

            $webResponse = Invoke-ClientRequest -Uri "apps/$AppName" -Accept "text/html"
            $LASTEXITCODE | Should -BeExactly 0
            $webResponse | Out-String | Should -BeLike "*Hi there. This is a test web application*"
        }

        AfterEach {
            PSc8y\Remove-Application -Id $AppName
        }
    }

    Context "existing application" {
        BeforeEach {
            $WebAppSource = "$PSScriptRoot/TestData/hosted-application/simple-helloworld"
            $AppName = New-RandomString -Prefix "app"
            $application = PSc8y\New-HostedApplication -File $WebAppSource -Name $AppName
        }

        It "Update an existing web application from a folder" {
            $application | Should -Not -BeNullOrEmpty
            $application.name | Should -BeExactly $AppName
            $application.contextPath | Should -BeExactly $AppName
            $application.resourcesUrl | Should -BeExactly "/"
            $application.activeVersionId | Should -MatchExactly "^\d+$"

            $applicationAfterUpdate = PSc8y\New-HostedApplication -File $WebAppSource -Name $AppName
            $LASTEXITCODE | Should -BeExactly 0
            $applicationAfterUpdate.activeVersionId | Should -MatchExactly "^\d+$"
            $applicationAfterUpdate.activeVersionId | Should -Not -BeExactly $application.activeVersionId

            $webResponse = Invoke-ClientRequest -Uri "apps/$AppName" -Accept "text/html"
            $webResponse | Out-String | Should -BeLike "*Hi there. This is a test web application*"
        }

        It -Tag "TODO" "Uploads new application but does not activate it" {
            $application = Get-Application -Id $AppName
            $application | Should -Not -BeNullOrEmpty
            $application.name | Should -BeExactly $AppName
            $application.contextPath | Should -BeExactly $AppName
            $application.resourcesUrl | Should -BeExactly "/"
            $application.activeVersionId | Should -MatchExactly "^\d+$"

            $applicationAfterUpdate = PSc8y\New-HostedApplication -File $WebAppSource -Name $AppName -SkipActivation
            $LASTEXITCODE | Should -BeExactly 0
            $applicationAfterUpdate.activeVersionId | Should -MatchExactly "^\d+$"
            $applicationAfterUpdate.activeVersionId | Should -BeExactly $application.activeVersionId

            $webResponse = Invoke-ClientRequest -Uri "apps/$AppName" -Accept "text/html"
            $webResponse | Out-String | Should -BeLike "*Hi there. This is a test web application*"
        }

        AfterEach {
            PSc8y\Remove-Application -Id $AppName
        }
    }
}
