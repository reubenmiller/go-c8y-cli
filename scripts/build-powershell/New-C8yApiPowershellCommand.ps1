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

        [string] $OutputDir = "./",

        [string] $DocBaseURL = "https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y"
    )

    $CmdletName = $Specification.alias.powershell
    $NounLower = $Noun.ToLower() -replace "/", " "
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
                Description = $iExample.description -replace "`"", "`'"
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

    # Link go command
    $null = $CmdletDocStringBuilder.AppendLine(".LINK")
    $pageName = "${NounLower}_${Verb}" -replace " ", "_"
    $null = $CmdletDocStringBuilder.AppendLine("$DocBaseURL/$pageName`n")
    # $NativeCommandName = @("c8y", $NounLower, $Verb) -join " "
    # $null = $CmdletDocStringBuilder.AppendLine("# [$NativeCommandName]($DocBaseURL/$pageName)`n")

    #
    # Arguments
    #
    $ArgumentSources = New-Object System.Collections.ArrayList

    if ($Specification.pathParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.pathParameters))
    }

    if ($Specification.queryParameters) {
        foreach ($item in $Specification.queryParameters) {
            if ($item.children) {
                # Ignore the item, and only use the children to build the cli arguments
                $null = $ArgumentSources.AddRange(([array]$item.children | Where-Object {
                    $_.type -ne "stringStatic"
                }))
            } else {
                $null = $ArgumentSources.Add($item)
            }
        }
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
    $C8yCommonSetNames = New-Object System.Collections.ArrayList

    $BeginParameterBuilder = New-Object System.Text.StringBuilder
    $ProcessParameterBuilder = New-Object System.Text.StringBuilder

    $null = $BeginParameterBuilder.AppendLine('        $Parameters = @{}')

    # Set iterator type
    $IteratorType = ""
    $IteratorVariable = ""
    $PipelineTemplateFormat = ""
    $NoEnumerate = $false

    # Sort argument sources by position (if specified)
    [array] $ArgumentSources = $ArgumentSources | ForEach-Object {
        if ($null -eq $_.position) {
            # Assign default value
            $_ | Add-Member -MemberType NoteProperty -Name position -Value 20
        }
        $_
    }

    # (stable) sort argument sources by position to control the expected order on cli
    [array] $ArgumentSources = $ArgumentSources `
    | Sort-Object -Property position -Stable `
    | Where-Object { -Not $_.skip } `
    | Where-Object { $_.name -ne "data" }

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

        if ($item.ignore) {
            Write-Warning "Skipping special argument: $($item.Name)"
            continue
        }

        if ($ReadFromPipeline) {
            Write-Host "$NameCamel : $($item.Name), $($iArg.type)=>$($item.type)" -ForegroundColor Magenta
            if ($item.type -match "^(string|long)$") {
                $item.type = "object[]"
            }
            $IteratorVariable = "`${0}" -f $item.Name
            $IteratorType = $item.type
        }

        # Set parameters
        if ($iArg.pipeline) {
            if ($iArg.required) {
                $null = $ProcessParameterBuilder.AppendLine("            if (`$item) {")
                $null = $ProcessParameterBuilder.AppendLine("                `$Parameters[`"$($iArg.Name)`"] = if (`$item.id) { `$item.id } else { `$item }")
                $null = $ProcessParameterBuilder.AppendLine("            }")
            } else {
                # $PipelineTemplateFormat = "loop_without_pipeline"
                # $ExpanderFunction = "PSc8y\Expand-Id `$$($item.Name)"
                # $null = $ProcessParameterBuilder.AppendLine("        `$Parameters[`"$($iArg.Name)`"] = $ExpanderFunction")
            }

        } else {

            $ItemValue = switch ($iArg.Type) {
                "switch" {
                    $true
                    break
                }

                "[]stringcsv" {
                    "`${0} -join ','" -f $item.Name
                    break
                }

                "hashtable" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                    break
                }

                "json" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                    break
                }

                "json_custom" {
                    "ConvertTo-JsonArgument `${0}" -f $item.Name
                    break
                }

                { $_ -match "agent|device|devicegroup|usergroup|role|user" } {
                    "PSc8y\Expand-Id `${0}" -f $item.Name
                    break
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

        if ($iArg.alias) {
            $null = $CurrentParam.AppendLine("        [Alias(`"{0}`")]" -f $iArg.alias)
        }

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
    # Add common parameters related to Method
    #
    switch ($Specification.method) {
        "GET" { $null = $C8yCommonSetNames.Add("Get") }
        "POST" {
            $null = $C8yCommonSetNames.Add("Create")
            $null = $C8yCommonSetNames.Add("Template")
        }
        "PUT" {
            $null = $C8yCommonSetNames.Add("Update")
            $null = $C8yCommonSetNames.Add("Template")
        }
        "DELETE" { $null = $C8yCommonSetNames.Add("Delete") }
    }

    #
    # Add common parameters
    #
    if ($ResultType -match "collection") {
        $null = $C8yCommonSetNames.Add("Collection")

        # note: don't enumerate collection output 
        $NoEnumerate = $false
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
    # Body
    #
    $RESTBodyBuilder = New-Object System.Text.StringBuilder
    if ($Specification.body) {
        $null = $RESTBodyBuilder.AppendLine('$body = @{}')

        foreach ($iArg in $Specification.body) {
            $prop = $iArg.property
            $type = $iArg.type

            if (!$prop) {
                $prop = $iArg.name
            }

            if ($prop) {
                if ($prop.Contains(".")) {
                    [array] $propParts = $prop -split "\."

                    # if ($propParts.Count -gt 2) {
                    #     Write-Warning "TODO: handle nested properties with depth > 2"
                    #     continue
                    # }
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
    # SupportsShouldProcess = `$true,
    # ConfirmImpact = '$CmdletConfirmImpact',
    $Template = @"
# Code generated from specification version 1.0.0: DO NOT EDIT
Function $CmdletName {
<#
$($CmdletDocStringBuilder.ToString())
#>
    [cmdletbinding(PositionalBinding=`$true,
                   HelpUri='$DocumentationLink')]
    [Alias()]
    [OutputType([object])]
    Param(
$($CmdletParameters -join ",`n`n")
    )
    DynamicParam {
        $(
            if ($null -ne $C8yCommonSetNames) {
                "Get-ClientCommonParameters -Type `"$($C8yCommonSetNames -join '", "')`""
            }
        )
    }

    Begin {
$(
    #$BeginParameterBuilder -join "`n"
)
        if (`$env:C8Y_DISABLE_INHERITANCE -ne `$true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet `$PSCmdlet -SessionState `$ExecutionContext.SessionState
        }

        `$c8yargs = New-ClientArgument -Parameters `$PSBoundParameters -Command "$NounLower $Verb"
        `$ClientOptions = Get-ClientOutputOption `$PSBoundParameters
        `$TypeOptions = @{
            Type = "$ResultType"
            ItemType = "$ResultItemType"
            BoundParameters = `$PSBoundParameters
        }
    }

    Process {
$(New-Body2 `
    -Noun $Noun `
    -PipelineTemplateFormat $PipelineTemplateFormat `
    -ConfirmImpact $CmdletConfirmImpact `
    -IteratorVariable $IteratorVariable `
    -SetParameters $ProcessParameterBuilder `
    -Verb $Verb `
    -IteratorType $IteratorType `
    -ResultType $ResultType `
    -ResultItemType $ResultItemType `
    -ResultSelectProperty $ResultSelectProperty `
    -NoEnumerate:$NoEnumerate)
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
        [string] $ConfirmImpact,
        [switch] $NoEnumerate
    )
    $NounLower = $Noun.ToLower() -replace "/", " "

    $ExpandFunction = Get-IteratorFunction -Type $IteratorType -Variable $IteratorVariable

    $ConfirmationStatement = ""
    $EnablePowershellConfirmation = $false
    if ($EnablePowershellConfirmation -and $ConfirmImpact -ne "None") {
        $ConfirmationStatement = @"
        `$Force = if (`$PSBoundParameters.ContainsKey("Force")) { `$PSBoundParameters["Force"] } else { `$False }
        if (!`$Force -and !`$WhatIfPreference) {
            `$items = $ExpandFunction

            `$shouldContinue = `$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name `$PSCmdlet.MyInvocation.InvocationName -InputObject `$items)
            )
            if (!`$shouldContinue) {
                return
            }
        }

"@
    }

    if ($NoEnumerate) {
        $prefix = ",("
        $suffix = ")"
    }

    $Template1 = @"
$ConfirmationStatement
$(
            if ($IteratorVariable) {
                @"
        if (`$ClientOptions.ConvertToPS) {
            ${prefix}$IteratorVariable ``
            | Group-ClientRequests ``
            | c8y $NounLower $Verb `$c8yargs ``
            | ConvertFrom-ClientOutput @TypeOptions${suffix}
        }
        else {
            $IteratorVariable ``
            | Group-ClientRequests ``
            | c8y $NounLower $Verb `$c8yargs
        }
        
"@
            } else {
            @"
        if (`$ClientOptions.ConvertToPS) {
            ${prefix}c8y $NounLower $Verb `$c8yargs ``
            | ConvertFrom-ClientOutput @TypeOptions${suffix}
        }
        else {
            c8y $NounLower $Verb `$c8yargs
        }
"@
            }
        )
"@
        $Template2 = @"
$ConfirmationStatement
$SetParameters
        if (`$ClientOptions.ConvertToPS) {
            ${prefix}c8y $NounLower $Verb `$c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions${suffix}
        }
        else {
            c8y $NounLower $Verb `$c8yargs
        }
"@


        # Return the appropriate process block
        switch ($PipelineTemplateFormat) {
            "loop_without_pipeline" { $Template2; break }
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
        "[]id" { "(PSc8y\Expand-Id $Variable)" }
        "[]role" { "(PSc8y\Expand-Id $Variable)" }
        "[]roleself" { "(PSc8y\Expand-Id $Variable)" }
        "[]tenant" { "(PSc8y\Expand-Tenant $Variable)" }
        "[]userself" { "(PSc8y\Expand-User $Variable)" }
        "[]user" { "(PSc8y\Expand-User $Variable)" }
        "application" { "(PSc8y\Expand-Application $Variable)" }
        "hostedapplication" { "(PSc8y\Expand-Application $Variable)" }
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
