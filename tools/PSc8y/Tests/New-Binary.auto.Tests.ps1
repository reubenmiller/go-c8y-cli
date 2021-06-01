. $PSScriptRoot/imports.ps1

Describe -Name "New-Binary" {
    BeforeEach {
        $File = New-TestFile
        $FileName = (Get-Item $File).Name

    }

    It "Upload a log file" {
        $Response = PSc8y\New-Binary -File $File
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Upload a config file and make it globally accessible for all users" {
        $Response = PSc8y\New-Binary -File $File -Type "c8y_upload" -Data @{ c8y_Global = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Item $File
        Find-ManagedObjectCollection -Query "has(c8y_IsBinary) and (name eq '$FileName')" | Remove-Binary

    }
}

