
. $PSScriptRoot/imports.ps1

Describe -Name "Get-SessionHomePath" {
    BeforeAll {
        $original = $env:C8Y_SESSION_HOME
    }

    BeforeEach {
        $tmpdir = New-TemporaryDirectory
    }

    It "Support a custom home session path" {
        $env:C8Y_SESSION_HOME = $tmpdir

        $c8yhome = Get-SessionHomePath
        $c8yhome | Should -BeExactly $env:C8Y_SESSION_HOME
    }

    It -Skip "Default to current directory if automatic HOME variable if it does not exist" {
        Set-Variable HOME ""
        $env:C8Y_SESSION_HOME = ""
        Set-Location $tmpdir

        $expectedPath = Join-Path "." -ChildPath ".cumulocity"
        $c8yhome = Get-SessionHomePath
        $c8yhome | Should -BeExactly $expectedPath
    }

    It -Skip "Use HOME automatic variable" {
        Set-Variable HOME "home/myuser"
        $env:C8Y_SESSION_HOME = ""

        $expectedPath = "home/myuser/.cumulocity"
        $c8yhome = Get-SessionHomePath
        $c8yhome | Should -BeExactly $expectedPath
    }

    AfterEach {
        if (Test-Path $tmpdir) {
            Remove-Item $tmpdir -Force -ErrorAction SilentlyContinue
        }
    }

    AfterAll {
        if ($original) {
            $env:C8Y_SESSION_HOME = $original
        }
    }
}
