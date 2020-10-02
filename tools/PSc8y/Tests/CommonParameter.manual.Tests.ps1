. $PSScriptRoot/imports.ps1

Describe -Name "Common parameters" {

    It "All commands support common c8y parameters" {
        $cmdlets = Get-Command -Module PSc8y -Name "*" | Where-Object {
            $_.Name -notmatch "Session"
        }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "OutputFile"
            $icmdlet | Should -HaveParameter "NoProxy"
            $icmdlet | Should -HaveParameter "Session"
            $icmdlet | Should -HaveParameter "TimeoutSec"
        }
    }

    It "All side-effect commands support WhatIf" {
        $ExcludeCmdlets = @(
            "Add-PowershellType",
            "New-RandomPassword",
            "New-RandomString",
            "New-Session",
            "Register-Alias",
            "Set-Session",
            "Set-ClientConsoleSetting"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" | Where-Object {
            $_.Name -match "^(Add|Update|Remove|Set|Reset|Register|New|Enable|Disable|Approve)"
        } | Where-Object {
            $ExcludeCmdlets -notcontains $_.name
        }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "WhatIf"
            $icmdlet | Should -HaveParameter "Force"
        }
    }

    It "All commands support Verbose" {
        $cmdlets = Get-Command -Module PSc8y -Name "*"

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "Verbose"
        }
    }
}
