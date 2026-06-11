. $PSScriptRoot/imports.ps1

Describe -Name "{{ CmdletName }}" {
    BeforeEach {
{{ BeforeEach }}
    }

{{ TestCases }}

    AfterEach {
{{ AfterEach }}
    }
}
