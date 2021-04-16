. $PSScriptRoot/../imports.ps1

Describe -Name "c8y filter common parameter" {
    BeforeAll {
        $TypeSuffix = New-RandomString
        $UniqueFragment = New-RandomString -Prefix "c8y_citest"
        $Template = "{ name: 'testFilter', type: 'c8yci_$TypeSuffix', ${UniqueFragment}: {}, values: [0, 1, 2], floatValue: 1.234, intValue: 100 }"
        $ids = New-Object System.Collections.ArrayList
        $Device1 = New-Device -Template $Template
        $Device2 = New-Device -Template $Template
        $null = $ids.AddRange(@($Device1.id, $Device2.id))
    }

    It "filters by wildcards" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "type like *$TypeSuffix*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "type -like *$TypeSuffix*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device1, $Device2
    }

    It "filters by negated wildcards" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "id notlike $($Device1.id)*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "id -notlike $($Device1.id)*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device2
    }

    It "filters by regex" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "type match c8yci_.+[a-z0-9]*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "type -match c8yci_.+[a-z0-9]*" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device1, $Device2
    }

    It "filters by negated regex" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "id notmatch $($Device1.id)x?" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "id -notmatch $($Device1.id)x?" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -ContainInCollection $Device2
    }

    It "filters by array length: greater than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lengt 0" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2
    }

    It "filters by array length: greater than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lengte 3" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lengte 4" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by array length: less than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lenlt 4" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lenlt 3" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by array length: less than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lenlte 3" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "values lenlte 2" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by int: greater than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gt 99" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gt 98.5000001" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gt 101" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gt 100.5" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by int: greater than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gte 100" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gte 99.1" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gte 101" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue gte 100.5" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by int: less than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lt 101" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lt 100.10001" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lt 99.99999" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lt 90" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by int: less than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lte 100" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lte 101.5" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lte -99.5" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "intValue lte 99.9999" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by float: greater than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue gt 1" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue gt 1.1" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue gt 2" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue gt 1.3" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by float: less than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue lt 2" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue lt 1.3" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue lt 1" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0

        $output = c8y devices list --fragmentType $UniqueFragment --filter "floatValue lt 1.1" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by string length: equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "name leneq 10" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "name leneq 11" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by string length: greater than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lengt 9" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lengt 10" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by string length: greater than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lengte 10" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lengte 11" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by string length: less than" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lenlt 11" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lenlt 10" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }

    It "filters by string length: less than or equal to" {
        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lenlte 10" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output | Should -ContainInCollection $Device1, $Device2

        $output = c8y devices list --fragmentType $UniqueFragment --filter "name lenlt 9" --orderBy "_id asc" | ConvertFrom-Json
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 0
    }
    
    AfterAll {
        $ids | Remove-ManagedObject
    }
}
