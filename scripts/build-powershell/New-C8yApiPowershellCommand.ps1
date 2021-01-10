Function New-C8yApiPowershellCommand {
    [cmdletbinding()]
    Param(
        [Parameter(
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Mandatory = $true
        )]
        [object[]] $Specification,

        [string] $Noun,

        [string] $OutputDir = "./"
    )

    $CmdletName = $Specification.alias.powershell
    $Name = $Specification.name
	$NameCamel = $Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)
    $File = Join-Path -Path $OutputDir -ChildPath ("{0}.ps1" -f $CmdletName)
    $ResultType = $Specification.accept
    $ResultItemType = $Specification.collectionType

    $ResultSelectProperty = $Specification.collectionProperty
    if ($ResultSelectProperty) {
        # Replace ".[]." syntax with "." as powershell handles nested access of arrays natively
        $ResultSelectProperty = $ResultSelectProperty -replace "\.#\.", "."
    }

    $Verb = $Specification.alias.go

    # Powershell Confirm impact
    $CmdletConfirmImpact = "None";
    if ($CmdletName -notmatch "^(Get|Find|Save)") {
        $CmdletConfirmImpact = "High"
    }

    #
    # Create test
    #
    if ($Specification.examples.powershell) {

        $TestCaseVariables = foreach ($iExample in $Specification.examples.powershell) {
            if (!$iExample.command) {
                continue
            }
            @{
                Command = $iExample.command
                Description = $iExample.description
                BeforeEach = $iExample.beforeEach
                AfterEach = $iExample.afterEach
            }
        }

        if ($TestCaseVariables) {

            # Adjust test case template depending if a response is expected or not
            $TestCaseTemplate = "$PSScriptRoot/templates/testcase.template.ps1"
            if ([string]::IsNullOrWhiteSpace($Specification.accept)) {
                $TestCaseTemplate = "$PSScriptRoot/templates/testcase.emptyresponse.template.ps1"
            }

            $SkipTest = $iExample.skipTest -eq $true

            New-C8yApiPowershellTest `
                -Name $CmdletName `
                -TestCaseVariables $TestCaseVariables `
                -OutFolder "$OutputDir/../Tests" `
                -SkipTest:$SkipTest `
                -TestCaseTemplateFile $TestCaseTemplate `
                -TemplateFile "$PSScriptRoot/templates/test.template.ps1"
        }
    }


    #
    # Meta information
    #
    $Synopsis = $Specification.description
    $DescriptionLong = $Specification.descriptionLong
    if (!$DescriptionLong) {
        $DescriptionLong = $Synopsis
    }
    $DocumentationLink = $Specification.link
    $Examples = foreach ($iExample in $Specification.examples.powershell) {
        if ($iExample.command) {
            $ExampleText = "`PS> {0}`n`n{1}" -f $iExample.command, $iExample.description
        } else {
            $ExampleText = $iExample
        }
        $ExampleText
    }

    $CmdletDocStringBuilder = New-Object System.Text.StringBuilder

    if ($Synopsis) {
        $null = $CmdletDocStringBuilder.AppendLine(".SYNOPSIS")
        $null = $CmdletDocStringBuilder.AppendLine("${Synopsis}`n")
    }

    if ($DescriptionLong) {
        $null = $CmdletDocStringBuilder.AppendLine(".DESCRIPTION")
        $null = $CmdletDocStringBuilder.AppendLine("${DescriptionLong}`n")
    }

    #
    # Arguments
    #
    $ArgumentSources = New-Object System.Collections.ArrayList

    if ($Specification.pathParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.pathParameters))
    }

    if ($Specification.queryParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.queryParameters))
    }

    if ($Specification.body) {
        $null = $ArgumentSources.AddRange(([array]$Specification.body))
    }

    if ($Specification.headerParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.headerParameters))
    }

    if ($Specification.options) {
        $null = $ArgumentSources.AddRange(([array]$Specification.options))
    }

    $CmdletParameters = New-Object System.Collections.ArrayList

    $BeginParameterBuilder = New-Object System.Text.StringBuilder
    $ProcessParameterBuilder = New-Object System.Text.StringBuilder

    $null = $BeginParameterBuilder.AppendLine('        $Parameters = @{}')

    # Set iterator type
    $IteratorType = ""
    $IteratorVariable = ""
    $PipelineTemplateFormat = ""

    # Sort argument sources by position (if specified)
    [array] $ArgumentSources = $ArgumentSources | ForEach-Object {
        if ($null -eq $_.position) {
            # Assign default value
            $_ | Add-Member -MemberType NoteProperty -Name position -Value 20
        }
        $_
    }

    # (stable) sort argument sources by position to control the expected order on cli
    [array] $ArgumentSources = $ArgumentSources |
        Sort-Object -Property position -Stable |
        Where-Object { -Not $_.skip }

    foreach ($iArg in $ArgumentSources) {
        $ReadFromPipeline = $iArg.pipeline -or $iArg.name -eq "id" -or $iArg.alias -eq "id"
        $ArgParams = @{
            Name = $iArg.name
            Type = $iArg.type
            OptionName = $iArg.alias
            Description = $iArg.description
            Default = $iArg.default
            Required = $iArg.required
            ReadFromPipeline = $ReadFromPipeline
        }
        $item = New-C8yPowershellArguments @ArgParams

        if ($ReadFromPipeline) {
            $IteratorVariable = "`${0}" -f $item.Name
            $IteratorType = $iArg.type
        }

        # Set parameters
        if ($iArg.pipeline) {
            if ($iArg.required) {
                $null = $ProcessParameterBuilder.AppendLine("            if (`$item) {")
                $null = $ProcessParameterBuilder.AppendLine("                `$Parameters[`"$($iArg.Name)`"] = if (`$item.id) { `$item.id } else { `$item }")
                $null = $ProcessParameterBuilder.AppendLine("            }")
            } else {
                $PipelineTemplateFormat = "loop_without_pipeline"
                $ExpanderFunction = "PSc8y\Expand-Id `$$($item.Name)"
                $null = $ProcessParameterBuilder.AppendLine("        `$Parameters[`"$($iArg.Name)`"] = $ExpanderFunction")
            }

        } else {

            $ItemValue = switch ($iArg.Type) {
                "switch" {
                    $true
                }

                "hashtable" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                }

                "json" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                }

                "json_custom" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                }

                "[]usergroup" {
                    "PSc8y\Expand-Id `${0}" -f $item.Name
                }

                default {
                    "`${0}" -f $item.Name
                }
            }
            $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"$($item.Name)`")) {")
            $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"$($iArg.Name)`"] = $ItemValue")
            $null = $BeginParameterBuilder.AppendLine("        }")
        }

        # Parameter definition
        $CurrentParam = New-Object System.Text.StringBuilder
        $null = $CurrentParam.AppendLine("        # {0}" -f ($item.Description))
        $null = $CurrentParam.AppendLine("        [Parameter({0})]" -f ($item.Definition -join ",`n                   "))

        # if ($iArg.alias) {
        #     $null = $CurrentParam.AppendLine("        [Alias(`"{0}`")]" -f $iArg.alias)
        # }

        # Validate set
        if ($null -ne $iArg.validationSet) {
            [array] $ValidationSet = $iArg.validationSet | Foreach-Object { "'$_'" }
            $null = $CurrentParam.AppendLine('        [ValidateSet({0})]' -f ($ValidationSet -join ","))
        }

        $null = $CurrentParam.AppendLine('        [{0}]' -f $item.Type)
        $null = $CurrentParam.Append('        ${0}' -f $item.Name)
        $null = $CmdletParameters.Add($CurrentParam)
    }

    #
    # Processing Mode
    #
    if ($Specification.method -match "DELETE|PUT|POST") {
        $ProcessingModeParam = New-Object System.Text.StringBuilder
        $null = $ProcessingModeParam.AppendLine('        # Cumulocity processing mode')
        $null = $ProcessingModeParam.AppendLine('        [Parameter()]')
        $null = $ProcessingModeParam.AppendLine('        [AllowNull()]')
        $null = $ProcessingModeParam.AppendLine('        [AllowEmptyString()]')
        $null = $ProcessingModeParam.AppendLine('        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "")]')
        $null = $ProcessingModeParam.AppendLine('        [string]')
        $null = $ProcessingModeParam.Append('        $ProcessingMode')
        $null = $CmdletParameters.Add($ProcessingModeParam)

        $null = $BeginParameterBuilder.AppendLine('        if ($PSBoundParameters.ContainsKey("ProcessingMode")) {')
        $null = $BeginParameterBuilder.AppendLine('            $Parameters["processingMode"] = $ProcessingMode')
        $null = $BeginParameterBuilder.AppendLine('        }')
    }

    #
    # Add common parameters
    #
    if ($ResultType -match "collection") {
        $PageSizeParam = New-Object System.Text.StringBuilder
        $null = $PageSizeParam.AppendLine('        # Maximum number of results')
        $null = $PageSizeParam.AppendLine('        [Parameter()]')
        $null = $PageSizeParam.AppendLine('        [AllowNull()]')
        $null = $PageSizeParam.AppendLine('        [AllowEmptyString()]')
        $null = $PageSizeParam.AppendLine('        [ValidateRange(1,2000)]')
        $null = $PageSizeParam.AppendLine('        [int]')
        $null = $PageSizeParam.Append('        $PageSize')
        $null = $CmdletParameters.Add($PageSizeParam)

        $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"PageSize`")) {")
        $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"pageSize`"] = `$PageSize")
        $null = $BeginParameterBuilder.AppendLine("        }")

        # If included, then the original data set will be returned
        $WithTotalPagesParam = New-Object System.Text.StringBuilder
        $null = $WithTotalPagesParam.AppendLine('        # Include total pages statistic')
        $null = $WithTotalPagesParam.AppendLine('        [Parameter()]')
        $null = $WithTotalPagesParam.AppendLine('        [switch]')
        $null = $WithTotalPagesParam.Append('        $WithTotalPages')
        $null = $CmdletParameters.Add($WithTotalPagesParam)

        # CurrentPage
        $CurrentPageParam = New-Object System.Text.StringBuilder
        $null = $CurrentPageParam.AppendLine('        # Get a specific page result')
        $null = $CurrentPageParam.AppendLine('        [Parameter()]')
        $null = $CurrentPageParam.AppendLine('        [int]')
        $null = $CurrentPageParam.Append('        $CurrentPage')
        $null = $CmdletParameters.Add($CurrentPageParam)

        # TotalPages: maximum number of pages to retreive when retrieving all pages
        $TotalPagesParam = New-Object System.Text.StringBuilder
        $null = $TotalPagesParam.AppendLine('        # Maximum number of pages to retrieve when using -IncludeAll')
        $null = $TotalPagesParam.AppendLine('        [Parameter()]')
        $null = $TotalPagesParam.AppendLine('        [int]')
        $null = $TotalPagesParam.Append('        $TotalPages')
        $null = $CmdletParameters.Add($TotalPagesParam)

        $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"WithTotalPages`") -and `$WithTotalPages) {")
        $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"withTotalPages`"] = `$WithTotalPages")
        $null = $BeginParameterBuilder.AppendLine("        }")

        #
        # Include option to expand pagination results
        # TODO: implement pagination results expansion in go
        $IncludeAllParam = New-Object System.Text.StringBuilder
        $null = $IncludeAllParam.AppendLine('        # Include all results')
        $null = $IncludeAllParam.AppendLine('        [Parameter()]')
        $null = $IncludeAllParam.AppendLine('        [switch]')
        $null = $IncludeAllParam.Append('        $IncludeAll')
        $null = $CmdletParameters.Add($IncludeAllParam)
    }

    # Template parameters (only for PUT)
    if ($Specification.Method -match "PUT|POST") {
        #
        # Template
        #
        $TemplateParam = New-Object System.Text.StringBuilder
        $null = $TemplateParam.AppendLine('        # Template (jsonnet) file to use to create the request body.')
        $null = $TemplateParam.AppendLine('        [Parameter()]')
        $null = $TemplateParam.AppendLine('        [string]')
        $null = $TemplateParam.Append('        $Template')
        $null = $CmdletParameters.Add($TemplateParam)

        $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"Template`") -and `$Template) {")
        $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"template`"] = `$Template")
        $null = $BeginParameterBuilder.AppendLine("        }")

        #
        # Template Variables
        #
        $TemplateVarsParam = New-Object System.Text.StringBuilder
        $null = $TemplateVarsParam.AppendLine('        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"')
        $null = $TemplateVarsParam.AppendLine('        [Parameter()]')
        $null = $TemplateVarsParam.AppendLine('        [string]')
        $null = $TemplateVarsParam.Append('        $TemplateVars')
        $null = $CmdletParameters.Add($TemplateVarsParam)

        $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"TemplateVars`") -and `$TemplateVars) {")
        $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"templateVars`"] = `$TemplateVars")
        $null = $BeginParameterBuilder.AppendLine("        }")
    }

    $RawParam = New-Object System.Text.StringBuilder
    $null = $RawParam.AppendLine('        # Show the full (raw) response from Cumulocity including pagination information')
    $null = $RawParam.AppendLine('        [Parameter()]')
    $null = $RawParam.AppendLine('        [switch]')
    $null = $RawParam.Append('        $Raw')
    $null = $CmdletParameters.Add($RawParam)

    $OutputFileParam = New-Object System.Text.StringBuilder
    $null = $OutputFileParam.AppendLine('        # Write the response to file')
    $null = $OutputFileParam.AppendLine('        [Parameter()]')
    $null = $OutputFileParam.AppendLine('        [string]')
    $null = $OutputFileParam.Append('        $OutputFile')
    $null = $CmdletParameters.Add($OutputFileParam)

    $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"OutputFile`")) {")
    $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"outputFile`"] = `$OutputFile")
    $null = $BeginParameterBuilder.AppendLine("        }")

    # No Proxy
    $NoProxyParam = New-Object System.Text.StringBuilder
    $null = $NoProxyParam.AppendLine('        # Ignore any proxy settings when running the cmdlet')
    $null = $NoProxyParam.AppendLine('        [Parameter()]')
    $null = $NoProxyParam.AppendLine('        [switch]')
    $null = $NoProxyParam.Append('        $NoProxy')
    $null = $CmdletParameters.Add($NoProxyParam)

    $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"NoProxy`")) {")
    $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"noProxy`"] = `$NoProxy")
    $null = $BeginParameterBuilder.AppendLine("        }")


    $SessionParam = New-Object System.Text.StringBuilder
    $null = $SessionParam.AppendLine('        # Specifiy alternative Cumulocity session to use when running the cmdlet')
    $null = $SessionParam.AppendLine('        [Parameter()]')
    $null = $SessionParam.AppendLine('        [string]')
    $null = $SessionParam.Append('        $Session')
    $null = $CmdletParameters.Add($SessionParam)

    $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"Session`")) {")
    $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"session`"] = `$Session")
    $null = $BeginParameterBuilder.AppendLine("        }")

    # Timeout (in seconds)
    $TimeoutSecParam = New-Object System.Text.StringBuilder
    $null = $TimeoutSecParam.AppendLine('        # TimeoutSec timeout in seconds before a request will be aborted')
    $null = $TimeoutSecParam.AppendLine('        [Parameter()]')
    $null = $TimeoutSecParam.AppendLine('        [double]')
    $null = $TimeoutSecParam.Append('        $TimeoutSec')
    $null = $CmdletParameters.Add($TimeoutSecParam)

    $null = $BeginParameterBuilder.AppendLine("        if (`$PSBoundParameters.ContainsKey(`"TimeoutSec`")) {")
    $null = $BeginParameterBuilder.AppendLine("            `$Parameters[`"timeout`"] = `$TimeoutSec * 1000")
    $null = $BeginParameterBuilder.AppendLine("        }")

    # Force parameter
    if ($Specification.method -match "(POST|PUT|DELETE)") {
        $ForceParam = New-Object System.Text.StringBuilder
        $null = $ForceParam.AppendLine("        # Don't prompt for confirmation")
        $null = $ForceParam.AppendLine('        [Parameter()]')
        $null = $ForceParam.AppendLine('        [switch]')
        $null = $ForceParam.Append('        $Force')
        $null = $CmdletParameters.Add($ForceParam)
    }

    # Examples
    foreach ($iExample in $Examples) {
        $null = $CmdletDocStringBuilder.AppendLine(".EXAMPLE")
        $null = $CmdletDocStringBuilder.AppendLine("${iExample}`n")
    }

    # Doc link
    if ($DocumentationLink) {
        $null = $CmdletDocStringBuilder.AppendLine(".LINK " + $DocumentationLink)
    }

    #
    #
    #

    $CmdletRestMethod = $Specification.method
    $CmdletRestPath = $Specification.path

    #
    # Body
    #
    $RESTBodyBuilder = New-Object System.Text.StringBuilder
    if ($Specification.body) {
        $null = $RESTBodyBuilder.AppendLine('$body = @{}')

        foreach ($iArg in $Specification.body) {
            $argname = $iArg.name
            $prop = $iArg.property
            $type = $iArg.type

            if (!$prop) {
                $prop = $iArg.name
            }

            if ($prop) {
                if ($prop.Contains(".")) {
                    [array] $propParts = $prop -split "\."

                    if ($propParts.Count -gt 2) {
                        Write-Warning "TODO: handle nested properties with depth > 2"
                        continue
                    }
                    $rootprop = $propParts[0]
                    $nestedprop = $propParts[1]
                    $null = $RESTBodyBuilder.AppendLine("`$body[`"$rootprop`"] = @{`"`" = `"$nestedprop`"}")
                } else {
                    switch ($type) {
                        "json" {
                            # Do nothing as it is already covered by getDataFlag
                        }
                        default {
                            $null = $RESTBodyBuilder.AppendLine("`$body[`"$prop`"] = ")
                        }
                    }
                }
            }
        }
    }

    #
    # Path Parameters
    #
    $RESTPathBuilder = New-Object System.Text.StringBuilder
    foreach ($iPathParameter in $Specification.pathParameters) {
        $prop = $iPathParameter.name
        $null = $RESTPathBuilder.AppendLine("")
    }

    #
    # Query parameters
    #
    $RESTQueryBuilder = New-Object System.Text.StringBuilder
    if ($Specification.queryParameters) {
        foreach ($iQueryParameter in $Specification.queryParameters) {
            $prop = $iQueryParameter.name
            $queryParam = $iQueryParameter.property
            if (!$queryParam) {
                $queryParam = $iQueryParameter.name
            }

            switch ($iQueryParameter.type) {
                "boolean" {
                    $null = $RESTQueryBuilder.AppendLine("")
                }

                "[]device" {
                    $null = $RESTQueryBuilder.AppendLine("")
                }

                # Array of strings
                "[]string" {
                    $null = $RESTQueryBuilder.AppendLine("")
                }

                default {
                    $null = $RESTQueryBuilder.AppendLine("")
                }
            }
        }
    }

    #
    # Template
    #
    $Template = @"
# Code generated from specification version 1.0.0: DO NOT EDIT
Function $CmdletName {
<#
$($CmdletDocStringBuilder.ToString())
#>
    [cmdletbinding(SupportsShouldProcess = `$true,
                   PositionalBinding=`$true,
                   HelpUri='$DocumentationLink',
                   ConfirmImpact = '$CmdletConfirmImpact')]
    [Alias()]
    [OutputType([object])]
    Param(
$($CmdletParameters -join ",`n`n")
    )

    Begin {
$($BeginParameterBuilder -join "`n")
    }

    Process {
$(New-Body2 -Noun $Noun -PipelineTemplateFormat $PipelineTemplateFormat -ConfirmImpact $CmdletConfirmImpact -IteratorVariable $IteratorVariable -SetParameters $ProcessParameterBuilder -Verb $Verb -IteratorType $IteratorType -ResultType $ResultType -ResultItemType $ResultItemType -ResultSelectProperty $ResultSelectProperty)
    }

    End {}
}
"@

	# Write to file with BOM (to help with encoding in powershell)
    $Encoding = New-Object System.Text.UTF8Encoding $true
	[System.IO.File]::WriteAllLines($File, $Template, $Encoding)
}

Function New-Body2 {
    Param(
        [string] $Noun,
        [string] $Verb,
        [string] $SetParameters,
        [string] $ResultType,
        [string] $ResultItemType,
        [string] $ResultSelectProperty,
        [string] $IteratorType,
        [string] $IteratorVariable,
        [string] $PipelineTemplateType,
        [string] $ConfirmImpact
    )

    $Target = "(PSc8y\Get-C8ySessionProperty -Name `"tenant`")"
    $Message = "(Format-ConfirmationMessage -Name `$PSCmdlet.MyInvocation.InvocationName -InputObject `$item)"

    $ExpandFunction = Get-IteratorFunction -Type $IteratorType -Variable $IteratorVariable

    $ConfirmationStatement = ""
    if ($ConfirmImpact -ne "None") {
        $ConfirmationStatement = @"
            if (!`$Force -and
                !`$WhatIfPreference -and
                !`$PSCmdlet.ShouldProcess(
                    $Target,
                    $Message
                )) {
                continue
            }

"@
    }

    $AdditionalArgs = ""
    if ($ResultType -match "Collection") {
        $AdditionalArgs = @"
 ``
                -CurrentPage:`$CurrentPage ``
                -TotalPages:`$TotalPages ``
                -IncludeAll:`$IncludeAll
"@.TrimEnd()
    }

    $Template1 = @"
        foreach (`$item in $ExpandFunction) {
$SetParameters
$ConfirmationStatement
            Invoke-ClientCommand ``
                -Noun "$Noun" ``
                -Verb "$Verb" ``
                -Parameters `$Parameters ``
                -Type "$ResultType" ``
                -ItemType "$ResultItemType" ``
                -ResultProperty "$ResultSelectProperty" ``
                -Raw:`$Raw${AdditionalArgs}
        }
"@
        $Template2 = @"
$SetParameters
        if (!`$Force -and
            !`$WhatIfPreference -and
            !`$PSCmdlet.ShouldProcess(
                $Target,
                $Message
            )) {
            continue
        }

        Invoke-ClientCommand ``
            -Noun "$Noun" ``
            -Verb "$Verb" ``
            -Parameters `$Parameters ``
            -Type "$ResultType" ``
            -ItemType "$ResultItemType" ``
            -ResultProperty "$ResultSelectProperty" ``
            -Raw:`$Raw ``
            -CurrentPage:`$CurrentPage ``
            -TotalPages:`$TotalPages ``
            -IncludeAll:`$IncludeAll
"@
        # Return the appropriate process block
        switch ($PipelineTemplateFormat) {
            "loop_without_pipeline" { $Template2 }
            default {
                $Template1
            }
        }
}

Function Get-IteratorFunction {
    Param(
        [string] $Type,

        [string] $Variable
    )

    $ExpandFunction = switch ($Type) {
        "[]device" { "(PSc8y\Expand-Device $Variable)" }
        "[]role" { "(PSc8y\Expand-Id $Variable)" }
        "[]roleself" { "(PSc8y\Expand-Id $Variable)" }
        "[]tenant" { "(PSc8y\Expand-Tenant $Variable)" }
        "[]userself" { "(PSc8y\Expand-User $Variable)" }
        "[]user" { "(PSc8y\Expand-User $Variable)" }
        "application" { "(PSc8y\Expand-Application $Variable)" }
        "microservice" { "(PSc8y\Expand-Microservice $Variable)" }
        "device" { "(PSc8y\Expand-Device $Variable)" }
        "event" { "(PSc8y\Expand-Event $Variable)" }
        "id" { "(PSc8y\Expand-Id $Variable)" }
        "managedObject" { "(PSc8y\Expand-ManagedObject $Variable)" }
        "source" { "(PSc8y\Expand-Source $Variable)" }
        default {
            if ($Variable -eq "") {
                "@(`"`")"
            } else {
                "(PSc8y\Expand-Id $Variable)"
            }
        }
    }

    $ExpandFunction
}
