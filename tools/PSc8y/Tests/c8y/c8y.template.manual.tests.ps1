. $PSScriptRoot/../imports.ps1

Describe -Name "c8y template" {
    It "template should preservce double quotes" {
        $output = c8y template execute --template '{\"email\": \"he ll@ex ample.com\"}'
        $LASTEXITCODE | Should -Be 0
        $body = $output | ConvertFrom-Json
        $body.email | Should -MatchExactly "he ll@ex ample.com"
    }

    It "provides relative time functions" {
        $output = c8y template execute --template "{now: _.Now(), randomNow: _.Now('-' + _.Int(10,60) + 'd'), nowRelative: _.Now('-1h'), nowNano: _.NowNano(), nowNanoRelative: _.NowNano('-10d')}"
        $LASTEXITCODE | Should -Be 0
        $data = $output | ConvertFrom-Json
        Get-Date $data.now | Should -Not -BeNullOrEmpty
        Get-Date $data.randomNow | Should -Not -BeNullOrEmpty
        Get-Date $data.nowRelative | Should -Not -BeNullOrEmpty
        Get-Date $data.nowNano | Should -Not -BeNullOrEmpty
        Get-Date $data.nowNanoRelative | Should -Not -BeNullOrEmpty
    }

    It "provides random number generators" {
        $output = c8y template execute --template "{int: _.Int(), int2: _.Int(-20), int3: _.Int(-50,-59), float: _.Float(), float2: _.Float(10), float3: _.Float(40, 45)}"
        $LASTEXITCODE | Should -Be 0
        $data = $output | ConvertFrom-Json
        $data.int | Should -BeGreaterOrEqual 0
        $data.int | Should -BeLessThan 100

        $data.int2 | Should -BeGreaterOrEqual -20
        $data.int2 | Should -BeLessThan 0

        $data.int3 | Should -BeGreaterOrEqual -59
        $data.int3 | Should -BeLessThan -50

        $data.float | Should -BeGreaterOrEqual 0
        $data.float | Should -BeLessThan 1

        $data.float2 | Should -BeGreaterOrEqual 0
        $data.float2 | Should -BeLessThan 10

        $data.float3 | Should -BeGreaterOrEqual 40
        $data.float3 | Should -BeLessThan 45
    }

    It "combines explicit arguments with data and templates parameters" {
        $inputdata = @{self = "https://example.com"} | ConvertTo-Json -Compress
        $output = $inputdata | c8y operations create `
            --device 12345 `
            --data 'other="1"' `
            --template "{c8y_DownloadConfigFile: {url: input.value['self']}}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            c8y_DownloadConfigFile = @{
                url = "https://example.com"
            }
            deviceId = "12345"
            other = 1
        }
    }

    It "explicit arguments override values from data and templates" {
        $inputdata = @{self = "https://example.com"} | ConvertTo-Json -Compress
        $output = $inputdata | c8y operations create `
            --device "1111" `
            --data 'deviceId=\"2222\"' `
            --template "{deviceId: '3333'}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            deviceId = "1111"
        }
    }

    It "piped arguments do not override data values" {
        $inputdata = @{deviceId = "1111"} | ConvertTo-Json -Compress
        $output = $inputdata | c8y operations create `
            --data 'deviceId=\"2222\"' `
            --template "{deviceId: '3333'}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            deviceId = "2222"
        }
    }

    It "piped arguments overide template variables" {
        $inputdata = @{deviceId = "1111"} | ConvertTo-Json -Compress
        $output = $inputdata | c8y operations create `
            --template "{deviceId: '3333'}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            deviceId = "1111"
        }
    }

    It "provides a generic way to remap pipes values to property that will not be picked up" {
        $inputdata = @{deviceId = "1111"} | ConvertTo-Json -Compress
        $output = $inputdata `
        | c8y util show --select "tempID:deviceId" `
        | c8y operations create `
            --template "{deviceId: '3333'}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            deviceId = "3333"
        }
    }

    It "uses piped input inside the template" {
        $inputdata = @{deviceId = "1111"} | ConvertTo-Json -Compress
        $output = $inputdata `
        | c8y util show --select "tempID:deviceId" `
        | c8y operations create `
            --template "{deviceId: input.value.tempID}" `
            --dry `
            --dryFormat json

        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json
        $request.body | Should -MatchObject @{
            deviceId = "1111"
        }
    }

    It "provides a function to get the path and query from a full url" {
        $output = "https://example.com/test/me?value=test&value=1" `
        | c8y template execute --template "{input:: input.value, name: _.GetURLPath(input.value)}" `
        | ConvertFrom-Json

        $output.name | Should -BeExactly "/test/me?value=test&value=1"
    }

    It "provides a function to get the scheme and hostname from a full url" {
        $output = "https://example.com/test/me?value=test&value=1" `
        | c8y template execute --template "{input:: input.value, name: _.GetURLHost(input.value)}" `
        | ConvertFrom-Json

        $output.name | Should -BeExactly "https://example.com"
    }
}
