Function New-C8yApiGoCommand {
    [cmdletbinding()]
    Param(
        [Parameter(
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Mandatory = $true
        )]
        [object[]] $Specification,

        [string] $OutputDir = "./"
    )

    $Name = $Specification.name
	$NameCamel = $Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)
	$File = Join-Path -Path $OutputDir -ChildPath ("{0}Cmd.auto.go" -f $Name)


    #
    # Meta information
    #
    $Use = $Specification.alias.go
    $Description = $Specification.description
    $DescriptionLong = $Specification.descriptionLong
    $Examples = foreach ($iExample in $Specification.examples.go) {
        if ($iExample.command) {
            $ExampleText = "`$ {0}`n{1}" -f $iExample.command, $iExample.description
        } else {
            $ExampleText = $iExample
        }
        $ExampleText
    }

    #
    # Arguments
    #
    $ArgumentSources = New-Object System.Collections.ArrayList

    # Path parameters
    if ($Specification.pathParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.pathParameters))
    }

    # Query parameters
    if ($Specification.queryParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.queryParameters))
    }

    # Body parameters
    if ($Specification.body) {
        $null = $ArgumentSources.AddRange(([array]$Specification.body))
    }

    # Header parameters
    if ($Specification.headerParameters) {
        $null = $ArgumentSources.AddRange(([array]$Specification.headerParameters))
    }

    # Additional parameters used to control a function
    if ($Specification.options) {
        $null = $ArgumentSources.AddRange(([array]$Specification.options))
    }

    $CommandArgs = New-Object System.Collections.ArrayList
    $PipelineVariableName = ""
    $PipelineVariableRequired = "false"
    $PipelineVariableProperty = ""
    $PipelineVariableAliases = ""
    $collectionProperty = ""

    if ($Specification.collectionProperty) {
        $collectionProperty = $Specification.collectionProperty
    }

    $CompletionBuilderOptions = New-Object System.Text.StringBuilder
    foreach ($iArg in (Remove-SkippedParameters $ArgumentSources)) {
        if ($iArg.pipeline) {
            $PipelineVariableName = $iArg.Name
            $PipelineVariableRequired = if ($iArg.Required) {"true"} else {"false"}
            $PipelineVariableProperty = if ($iArg.Property) { $iArg.Property } else { $iArg.Name }
            $PipelineVariableAliases = $iArg.pipelineAliases
            if (!$PipelineVariableAliases) {
                if ($PipelineVariableName -match "device$" -or $iArg.type -match "device$") {
                    $PipelineVariableAliases = @(
                        "deviceId",
                        "source.id",
                        "managedObject.id",
                        "id"
                    )
                } elseif ($PipelineVariableName -ne "id") {
                    $PipelineVariableAliases = @("id")
                }
            }
        }
        if ($iArg.validationSet) {
            $validateSetOptions = @($iArg.validationSet | ForEach-Object { "`"$_`"" }) -join ","
            $null = $CompletionBuilderOptions.AppendLine("completion.WithValidateSet(`"$($iarg.Name)`", $validateSetOptions),")
            
        }
        $ArgParams = @{
            Name = $iArg.name
            Type = $iArg.type
            OptionName = $iArg.alias
            Description = $iArg.description
            Default = $iArg.default
            Required = $iArg.required
            Pipeline = $iArg.pipeline
        }
        $arg = Get-C8yGoArgs @ArgParams
        $null = $CommandArgs.Add($arg)
    }

    if (!$PipelineVariableName -and $ArgumentSources.Count -gt 0) {
        Write-Warning ("Property is missing pipeline support. cmd={0}" -f @(
            $Specification.name
        ))
    }

    # Add common parameters
    if ($Specification.method -match "DELETE|PUT|POST") {
        $null = $CommandArgs.Add(@{
            SetFlag = 'addProcessingModeFlag(cmd)'
        })
    }

    $RESTPath = $Specification.path
    $RESTMethod = $Specification.method

    #
    # Body
    #
    $RESTBodyBuilderOptions = New-Object System.Text.StringBuilder
    $RESTFormDataBuilderOptions = New-Object System.Text.StringBuilder
    $GetBodyContents = "body"
    
    if ($Specification.body) {
        switch ($Specification.bodyContent.type) {
            "binary" {
                $GetBodyContents = "body.GetFileContents()"
                break
            }
            "formdata" {
                $GetBodyContents = "body"
                break
            }
            default {
                $GetBodyContents = "body"
                $null = $RESTBodyBuilderOptions.AppendLine("WithDataValue(),")
            }
        }

        foreach ($iArg in (Remove-SkippedParameters $Specification.body)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "body"
            if ($code) {
                switch -Regex ($code) {
                    "^flags\.WithFormData" {
                        $null = $RESTFormDataBuilderOptions.AppendLine($code)
                        break
                    }

                    "^(flags\.|With)" {
                        $null = $RESTBodyBuilderOptions.AppendLine($code)
                        break
                    }

                    default {
                        Write-Warning "Unknown body code. $code"
                    }
                }
            } else {
                Write-Warning ("No setter found for [{0}]" -f $iArg.name)
            }
        }

        #
        # Activate seperate body templating (if not included in -Data parameter)
        #
        if ($Specification.bodyTemplateOptions.enabled -eq $true) {
            $CommandArgs += @{
                SetFlag = "addTemplateFlag(cmd)"
            }
        }

        #
        # Apply a body template to the data
        #
        if ($Specification.bodyTemplate) {
            switch ($Specification.bodyTemplate.type) {
                "jsonnet" {
                    # ApplyLast: true == apply template to the existing json (potentially overriding values)
                    #            false == Use template as base json, and the existing json will take precendence
                    if ($Specification.bodyTemplate.applyLast -eq "true") {
                        $null = $RESTBodyBuilderOptions.AppendLine("flags.WithRequiredTemplateString(```n{0}``)," -f @(
                            $Specification.bodyTemplate.template
                        ))
                        
                    } else {
                        $null = $RESTBodyBuilderOptions.AppendLine("flags.WithDefaultTemplateString(```n{0}``)," -f @(
                            $Specification.bodyTemplate.template
                        ))
                    }

                }
                "none" {
                    # Do nothing
                }
                default {
                    Write-Warning ("Unsupported templating type [{0}]" -f $Specification.bodyTemplate.type)
                }
            }
        }

        #
        # Add support for user defined templates to control body
        #
        if ($Specification.bodyTemplate.type -ne "none") {
            $null = $RESTBodyBuilderOptions.AppendLine("WithTemplateValue(),")
            $null = $RESTBodyBuilderOptions.AppendLine("WithTemplateVariablesValue(),")
        }

        if ($Specification.bodyRequiredKeys) {
            $literalValues = ($Specification.bodyRequiredKeys | Foreach-Object {
                '"{0}"' -f $_
            }) -join ", "
            $null = $RESTBodyBuilderOptions.AppendLine("flags.WithRequiredProperties({0})," -f $literalValues)
        }
    }

    #
    # Host
    #
    $RESTHost = ""
    if ($null -ne $Specification.host) {
        $RESTHost = "`nHost:         replacePathParameters(`"$($Specification.host)`", pathParameters),"
    }

    #
    # Path Parameters
    #
    $RESTPathBuilderOptions = New-Object System.Text.StringBuilder
    foreach ($Properties in (Remove-SkippedParameters $Specification.pathParameters)) {
        $code = New-C8yApiGoGetValueFromFlag -Parameters $Properties -SetterType "path"
        if ($code) {
            $null = $RESTPathBuilderOptions.AppendLine($code)
        }
    }

    #
    # Query parameters
    #
    $RESTQueryBuilderWithValues = New-Object System.Text.StringBuilder
    $RESTQueryBuilderPost = New-Object System.Text.StringBuilder
    if ($Specification.queryParameters) {
        foreach ($Properties in (Remove-SkippedParameters $Specification.queryParameters)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $Properties -SetterType "query"
            if ($code) {
                $null = $RESTQueryBuilderWithValues.AppendLine($code)
            }
        }
    }
    if ($Specification.method -match "GET") {
        
        $null = $RESTQueryBuilderPost.AppendLine(@"
        commonOptions, err := getCommonOptions(cmd)
        if err != nil {
            return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
        }
"@)
        $null = $RESTQueryBuilderPost.AppendLine("commonOptions.AddQueryParameters(query)")
    }

    #
    # Headers
    #
    $RestHeaderBuilderOptions = New-Object System.Text.StringBuilder
    if ($Specification.headerParameters) {
        foreach ($iArg in (Remove-SkippedParameters $Specification.headerParameters)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "header"
            if ($code) {
                $null = $RestHeaderBuilderOptions.AppendLine($code)
            }
        }
    }

    # Processing Mode
    if ($Specification.method -match "DELETE|PUT|POST") {
        $null = $RestHeaderBuilderOptions.AppendLine("flags.WithProcessingModeValue(),")
    }

    #
    # TODO: Check if this option can be removed
    # Options
    #
    $RESTOptionsBuilder = New-Object System.Text.StringBuilder
    if ($Specification.options) {
        $null = $RESTOptionsBuilder.AppendLine('body.SetMap(getDataFlag(cmd))')

        foreach ($iArg in $Specification.options) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "body"
            if ($code) {
                $null = $RESTOptionsBuilder.AppendLine($code)
            }
        }
    }


    #
    # Pre run validation (disable some commands without switch flags)
    #
    $PreRunFunction = switch ($Specification.method) {
        "POST" { "validateCreateMode" }
        "PUT" { "validateUpdateMode" }
        "DELETE" { "validateDeleteMode" }
        default { "nil" }
    }

    #
    # Template
    #
    $Template = @"
// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

    "github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
    "github.com/reubenmiller/go-c8y-cli/pkg/completion"
    "github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// ${NameCamel}Cmd command
type ${NameCamel}Cmd struct {
    *baseCmd
}

// New${NameCamel}Cmd creates a command to $Description
func New${NameCamel}Cmd() *${NameCamel}Cmd {
	ccmd := &${NameCamel}Cmd{}
	cmd := &cobra.Command{
		Use:   "$Use",
		Short: "$Description",
		Long:  ``$DescriptionLong``,
        Example: ``
$($Examples -join "`n`n")
        ``,
        PreRunE: $PreRunFunction,
		RunE: ccmd.RunE,
    }

    cmd.SilenceUsage = true

    $($CommandArgs.SetFlag -join "`n	")

    completion.WithOptions(
		cmd,
		$CompletionBuilderOptions
	)

    flags.WithOptions(
		cmd,
        $(
            if ($PipelineVariableAliases) {
                $aliases = ($PipelineVariableAliases | ForEach-Object { "`"$_`""` }) -join ", "
                "flags.WithExtendedPipelineSupport(`"$PipelineVariableName`", `"$PipelineVariableProperty`", $PipelineVariableRequired, $aliases),"
            } else {
                "flags.WithExtendedPipelineSupport(`"$PipelineVariableName`", `"$PipelineVariableProperty`", $PipelineVariableRequired),"
            }   
        )
        $(
            if ($collectionProperty) {
                "flags.WithCollectionProperty(`"$collectionProperty`"),"
            }
        )
	)

    // Required flags
    $($CommandArgs.Required -join "`n	")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *${NameCamel}Cmd) RunE(cmd *cobra.Command, args []string) error {
    var err error
    inputIterators, err := flags.NewRequestInputIterators(cmd)
    if err != nil {
        return err
    }

    // query parameters
    query := flags.NewQueryTemplate()
    err = flags.WithQueryParameters(
		cmd,
        query,
        inputIterators,
        $RESTQueryBuilderWithValues
    )
    if err != nil {
		return cmderrors.NewUserError(err)
    }
    $RESTQueryBuilderPost
	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

    // headers
    headers := http.Header{}
    err = flags.WithHeaders(
		cmd,
        headers,
        inputIterators,
        $RestHeaderBuilderOptions
    )
    if err != nil {
		return cmderrors.NewUserError(err)
    }

    // form data
    formData := make(map[string]io.Reader)
    err = flags.WithFormDataOptions(
		cmd,
        formData,
        inputIterators,
		$RESTFormDataBuilderOptions
    )
    if err != nil {
		return cmderrors.NewUserError(err)
    }
    

    // body
    body := mapbuilder.NewInitializedMapBuilder()
    err = flags.WithBody(
        cmd,
        body,
        inputIterators,
        $RESTBodyBuilderOptions
    )
    if err != nil {
		return cmderrors.NewUserError(err)
    }

    // path parameters
    path := flags.NewStringTemplate("${RESTPath}")
    err = flags.WithPathParameters(
        cmd,
        path,
        inputIterators,
        $RESTPathBuilderOptions
    )
    if err != nil {
        return err
    }

    req := c8y.RequestOptions{$RESTHost
        Method:       "${RESTMethod}",
        Path:         path.GetTemplate(),
        Query:        queryValue,
        Body:         $GetBodyContents,
        FormData:     formData,
        Header:       headers,
        IgnoreAccept: globalFlagIgnoreAccept,
        DryRun:       globalFlagDryRun,
    }

    return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}

"@

    # Must not include BOM!
	$Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
	[System.IO.File]::WriteAllLines($File, $Template, $Utf8NoBomEncoding)

	# Auto format code (using goimports as it removes unused imports)
	& goimports -w $File
}

Function Remove-SkippedParameters {
<#
.SYNOPSIS
Remove skipped parameters. These are parameter which should not be used when generating code.
#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [AllowEmptyCollection()]
        [AllowNull()]
        [object[]] $CommandParameters
    )

    $CommandParameters | Where-Object {
        if ($_.skip -eq $true) {
            Write-Verbose ("Skipping parameter [{0}] as it is marked as skip" -f $_.name)
        }
        $_.skip -ne $true
    }
}

Function Get-C8yGoArgs {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Name,

        [Parameter(
            Mandatory = $true,
            Position = 1
        )]
        [string] $Type,

        [string] $Required,

        [string] $OptionName,

        [string] $Description,

        [string] $Default,

        [string] $Pipeline
    )

    if ($Required -match "true|yes") {
        $Description = "${Description} (required)"
    }

    if ($Pipeline -match "true|yes") {
        $Description = "${Description} (accepts pipeline)"
    }

    $Entry = switch ($Type) {
        "id" {
            # TODO: refactor addIdFlag to accept a property name
            if ($Name -eq "id") {
                @{
                    SetFlag = "addIDFlag(cmd)"
                }
            } else {
                $SetFlag = if ($UseOption) {
                    'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
                } else {
                    'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
                }
                @{
                    SetFlag = $SetFlag
                }
            }
        }

        "json" {
            @{
                SetFlag = "addDataFlag(cmd)"
            }
        }

        #
        # Usage: Accept json, but assign it to a nested property
        #
        "json_custom" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        { @("datefrom", "dateto", "datetime") -contains $_ } {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "source" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "directory" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "[]string" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSlice(`"${Name}`", `"${OptionName}`", []string{`"${Default}`"}, `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "[]stringcsv" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSlice(`"${Name}`", `"${OptionName}`", []string{`"${Default}`"}, `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }
            @{
                SetFlag = $SetFlag
            }
            break
        }

        "[]device" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]agent" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]devicegroup" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]roleself" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]role" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]usergroup" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "application" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "microservice" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "string" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "integer" {
            try {
                $DefaultInt = [convert]::ToInt64($Default)
            } catch {
                $DefaultInt = 0
            }

            $SetFlag = if ($UseOption) {
                'cmd.Flags().IntP("{0}", "{1}", {2}, "{3}")' -f $Name, $OptionName, $DefaultInt, $Description
            } else {
                'cmd.Flags().Int("{0}", {1}, "{2}")' -f $Name, $DefaultInt, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "float" {
            try {
                $DefaultFloat = [convert]::ToDecimal($Default)
            } catch {
                $DefaultFloat = 0
            }

            $SetFlag = if ($UseOption) {
                'cmd.Flags().Float32P("{0}", "{1}", {2}, "{3}")' -f $Name, $OptionName, $DefaultFloat, $Description
            } else {
                'cmd.Flags().Float32("{0}", {1}, "{2}")' -f $Name, $DefaultFloat, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "tenant" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "file" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "attachment" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "fileContents" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "boolean" {
            if (!$Default) {
                $Default = "false"
            }
            $SetFlag = if ($UseOption) {
                'cmd.Flags().BoolP("{0}", "{1}", {2}, "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().Bool("{0}", {1}, "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]user" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "[]userself" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        default {
            Write-Warning "Unknown flag type [$_]"
        }
    }

    # Set required flag
    if ($Required -match "true|yes" -and $Pipeline -notmatch "true") {
        $Entry | Add-Member -MemberType NoteProperty -Name "Required" -Value "_ = cmd.MarkFlagRequired(`"${Name}`")"
        # $Entry.Required = "cmd.MarkFlagRequired(`"${Name}`")"
    }

    $Entry
}

