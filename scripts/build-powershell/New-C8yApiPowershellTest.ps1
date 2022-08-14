Function New-C8yApiPowershellTest {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Name,

        [Parameter(
            Mandatory = $true
        )]
        [hashtable[]] $TestCaseVariables,

        [string] $TestCaseTemplateFile,

        [switch] $SkipTest,

        [string] $TemplateFile = "templates/test.template.ps1",

        [Parameter(
            Mandatory = $true
        )]
        [string] $OutFolder,

        [switch] $Deprecated
    )

    $Template = Get-Content $TemplateFile -Raw

    $TestCaseTemplate = Get-Content $TestCaseTemplateFile -Raw

    $BeforeBlock = New-Object System.Text.StringBuilder
    $AfterBlock = New-Object System.Text.StringBuilder

    $TestCases = foreach ($TestCase in $TestCaseVariables) {
        $iTestCaseTemplate = "$TestCaseTemplate"

        # Skip Test
        if ($SkipTest) {
            $iTestCaseTemplate = $iTestCaseTemplate -replace '\bIt "', 'It -Skip "'
        }

        # Add any explicit before blocks
        if ($null -ne $TestCase.BeforeEach) {
            foreach ($statement in $TestCase.BeforeEach) {
                if (!$BeforeBlock.ToString().Contains("$statement")) {
                    $null = $BeforeBlock.AppendLine("        $statement")
                }
            }
        }

        # Add any explicit before blocks
        if ($null -ne $TestCase.AfterEach) {
            foreach ($statement in $TestCase.AfterEach) {
                if (!$AfterBlock.ToString().Contains("$statement")) {
                    $null = $AfterBlock.AppendLine("        $statement")
                }
            }
        }

        # Replace any random variables
        # TODO: Check if a test device has already been created already in the before block
        #       then the  -ErrorAction silentlycontinue can be removed.
        if ($TestCase.Command -match "{{\s*randomdevice\s*}}") {
            $BeforeStatement = '$TestDevice = PSc8y\New-TestDevice'
            if (!$BeforeBlock.ToString().Contains($BeforeStatement)) {
                $null = $BeforeBlock.AppendLine("        $BeforeStatement")
            }

            $TestCase.Command = $TestCase.Command -replace "`"?{{\s*randomdevice\s*}}`"?", "`$TestDevice.id"

            $AfterStatement = 'if ($TestDevice.id) {'
            if (!$AfterBlock.ToString().Contains($AfterStatement)) {
                $null = $AfterBlock.AppendLine("        $AfterStatement")
                $null = $AfterBlock.AppendLine("            PSc8y\Remove-ManagedObject -Id `$TestDevice.id -ErrorAction SilentlyContinue")
                $null = $AfterBlock.AppendLine("        }")
            }
        }

        if ($TestCase.Command -match "{{\s*randomagent\s*}}") {
            $null = $BeforeBlock.AppendLine("        `$TestAgent = PSc8y\New-TestAgent")

            $TestCase.Command = $TestCase.Command -replace "`"?{{\s*randomagent\s*}}`"?", "`$TestAgent.id"

            $null = $AfterBlock.AppendLine("        if (`$TestAgent.id) {")
            $null = $AfterBlock.AppendLine("            PSc8y\Remove-ManagedObject -Id `$TestAgent.id -ErrorAction SilentlyContinue")
            $null = $AfterBlock.AppendLine("        }")
        }

        if ($TestCase.Command -match "{{\s*NewAlarm\s*}}") {
            $null = $BeforeBlock.AppendLine("        `$TestAlarm = PSc8y\New-TestAlarm")

            $TestCase.Command = $TestCase.Command -replace "`"?{{\s*NewAlarm\s*}}`"?", "`$TestAlarm.id"

            $null = $AfterBlock.AppendLine("        if (`$TestAlarm.source.id) {")
            $null = $AfterBlock.AppendLine("            PSc8y\Remove-ManagedObject -Id `$TestAlarm.source.id -ErrorAction SilentlyContinue")
            $null = $AfterBlock.AppendLine("        }")
        }

        if ($TestCase.Command -match "{{\s*NewOperation\s*}}") {
            $null = $BeforeBlock.AppendLine("        `$TestOperation = PSc8y\New-TestOperation")

            $TestCase.Command = $TestCase.Command -replace "`"?{{\s*NewOperation\s*}}`"?", "`$TestOperation.id"

            $null = $AfterBlock.AppendLine("        if (`$TestOperation.deviceId) {")
            $null = $AfterBlock.AppendLine("            PSc8y\Remove-ManagedObject -Id `$TestOperation.deviceId -ErrorAction SilentlyContinue")
            $null = $AfterBlock.AppendLine("        }")
        }

        if ($TestCase.Command -match "{{\s*NewEvent\s*}}") {
            $null = $BeforeBlock.AppendLine("        `$TestEvent = PSc8y\New-TestEvent")

            $TestCase.Command = $TestCase.Command -replace "`"?{{\s*NewEvent\s*}}`"?", "`$TestEvent.id"

            $null = $AfterBlock.AppendLine("        if (`$TestEvent.source.id) {")
            $null = $AfterBlock.AppendLine("            PSc8y\Remove-ManagedObject -Id `$TestEvent.source.id -ErrorAction SilentlyContinue")
            $null = $AfterBlock.AppendLine("        }")
        }


        # Create test case
        foreach ($variableName in $TestCase.Keys) {
            $iTestCaseTemplate = $iTestCaseTemplate -replace "{{\s*$variableName\s*}}", $TestCase[$variableName]
        }
        $iTestCaseTemplate
    }

    $Variables = @{
        CmdletName = $Name
        TestCases = $TestCases -join "`n"
        BeforeEach = $BeforeBlock
        AfterEach = $AfterBlock
    }

    foreach ($variableName in $Variables.Keys) {
        $Template = $Template -replace "{{\s*$variableName\s*}}", $Variables[$variableName]
    }

    $OutFile = Join-Path -Path $OutFolder -ChildPath "${Name}.auto.Tests.ps1"

    if ($Deprecated) {
        Remove-Item -Path $OutFile -ErrorAction SilentlyContinue
    } else {
        # Write to file with BOM (to help with encoding in powershell)
        $Encoding = New-Object System.Text.UTF8Encoding $true
        [System.IO.File]::WriteAllLines($OutFile, $Template, $Encoding)
    }
}
