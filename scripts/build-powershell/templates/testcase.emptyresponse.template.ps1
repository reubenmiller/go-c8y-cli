    It "{{ Description }}" {
        $Response = PSc8y\{{ Command }}
        $LASTEXITCODE | Should -Be 0
    }
