﻿. $PSScriptRoot/imports.ps1

Describe -Name "Get-DataHubJob" {
    BeforeEach {

    }

    It -Skip "Retrieve a datahub job" {
        $Response = PSc8y\Get-DataHubJob -Id "22feee74-875a-561c-5508-04114bdda000"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

