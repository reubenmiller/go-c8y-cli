. $PSScriptRoot/imports.ps1

Describe -Name "New-Binary" {
    BeforeEach {
        $File = New-TestFile
        $FileName = (Get-Item $File).Name

    }

    It "Upload a log file with custom properties" {
        $Response = PSc8y\New-Binary -File $File -Data @{ type = "c8y_upload"; c8y_Global = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.id | Should -MatchExactly "^\d+$"
        $Response.type | Should -BeExactly "c8y_upload"
        $Response.c8y_Global | Should -BeTrue
    }

    It "Upload a log file with custom properties but let file type be detected" {
        $Response = PSc8y\New-Binary -File $File -Data @{ c8y_Global = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.id | Should -MatchExactly "^\d+$"
        $Response.type | Should -BeExactly "application/octet-stream"
        $Response.c8y_Global | Should -BeTrue
    }


    AfterEach {
        Remove-Item $File
        Find-ManagedObjectCollection -Query "has(c8y_IsBinary) and (name eq '$FileName')" | Remove-Binary

    }
}

