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
    foreach ($iArg in (Remove-SkippedParameters $ArgumentSources)) {
        $ArgParams = @{
            Name = $iArg.name
            Type = $iArg.type
            OptionName = $iArg.alias
            Description = $iArg.description
            Default = $iArg.default
            Required = $iArg.required
        }
        $arg = Get-C8yGoArgs @ArgParams
        $null = $CommandArgs.Add($arg)
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
    $RESTBodyBuilder = New-Object System.Text.StringBuilder
    $RESTFormDataBuilder = New-Object System.Text.StringBuilder
    $GetBodyContents = "body.GetMap()"
    
    if ($Specification.body) {
        if ($Specification.bodyContent.type -ne 'binary') {
            $null = $RESTBodyBuilder.AppendLine('body.SetMap(getDataFlag(cmd))')
        } else {
            $GetBodyContents = "body.GetFileContents()"
        }

        foreach ($iArg in (Remove-SkippedParameters $Specification.body)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "body"
            if ($code) {
                $null = $RESTBodyBuilder.AppendLine($code)
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
                    $Reverse = "true"
                    if ($Specification.bodyTemplate.applyLast -eq "true") {
                        $Reverse = "false"
                    }
                    $null = $RESTBodyBuilder.AppendLine("bodyErr := body.MergeJsonnet(```n{0}``, {1})" -f @(
                        $Specification.bodyTemplate.template,
                        $Reverse
                    ))

                    $BodyErrCheck = @"
        if bodyErr != nil {
            return newSystemError("Template error. ", bodyErr)
        }
"@.TrimStart()
                    $null = $RESTBodyBuilder.AppendLine($BodyErrCheck)
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
            $BodyUserTemplateCode = @"
        if err := setDataTemplateFromFlags(cmd, body); err != nil {
            return newUserError("Template error. ", err)
        }
"@.TrimStart()
            $null = $RESTBodyBuilder.AppendLine($BodyUserTemplateCode)
        }
        
        if ($Specification.bodyValidation) {
            switch ($Specification.bodyValidation.type) {
                "jsonnet" {
                    $null = $RESTBodyBuilder.AppendLine("body.SetValidateTemplate(```n{0}``)" -f $Specification.bodyValidation.template)
                }
                default {
                    Write-Warning ("Unsupported body validation template type [{0}]" -f $Specification.bodyValidation.type)
                }
            }
        }

        if ($Specification.bodyRequiredKeys) {
            $literalValues = ($Specification.bodyRequiredKeys | Foreach-Object {
                '"{0}"' -f $_
            }) -join ", "
            $null = $RESTBodyBuilder.AppendLine("body.SetRequiredKeys({0})" -f $literalValues)
        }

        #
        # Validate body
        #
        if ($Specification.bodyContent.type -ne 'binary') {
            $BodyValidateionCode = @"
        if err := body.Validate(); err != nil {
            return newUserError("Body validation error. ", err)
        }
"@.TrimStart()
            $null = $RESTBodyBuilder.AppendLine($BodyValidateionCode)
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
    $RESTPathBuilder = New-Object System.Text.StringBuilder
    foreach ($Properties in (Remove-SkippedParameters $Specification.pathParameters)) {
        $code = New-C8yApiGoGetValueFromFlag -Parameters $Properties -SetterType "path"
        if ($code) {
            $null = $RESTPathBuilder.AppendLine($code)
        }
    }

    #
    # Query parameters
    #
    $RESTQueryBuilder = New-Object System.Text.StringBuilder
    $null = $RESTQueryBuilder.AppendLine('query := url.Values{}')
    if ($Specification.queryParameters) {
        foreach ($Properties in (Remove-SkippedParameters $Specification.queryParameters)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $Properties -SetterType "query"
            if ($code) {
                $null = $RESTQueryBuilder.AppendLine($code)
            }
        }
    }
    if ($Specification.method -match "GET") {
        $null = $RESTQueryBuilder.AppendLine("commonOptions.AddQueryParameters(&query)")
    }

    #
    # Headers
    #
    $RestHeaderBuilder = New-Object System.Text.StringBuilder
    if ($Specification.headerParameters) {
        foreach ($iArg in (Remove-SkippedParameters $Specification.headerParameters)) {
            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "header"
            if ($code) {
                $null = $RestHeaderBuilder.AppendLine($code)
            }
        }
    }

    # Processing Mode
    if ($Specification.method -match "DELETE|PUT|POST") {
        $null = $RestHeaderBuilder.AppendLine(@"
     if cmd.Flags().Changed("processingMode") {
         if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
             headers.Add("X-Cumulocity-Processing-Mode", v)
         }
     }
"@)
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
    # Add common options
    #
#     $null = $RESTQueryBuilder.AppendLine(@"
#     if cmd.Flags().Changed("pageSize") {
#         if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
#             query.Add("pageSize", fmt.Sprintf("%d", v))
#         }
#     }
# "@)
    #
    # Encode query parameters to a string
    #
    $null = $RESTQueryBuilder.AppendLine(@"
    queryValue, err = url.QueryUnescape(query.Encode())

    if err != nil {
        return newSystemError("Invalid query parameter")
    }
"@)

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

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type ${Name}Cmd struct {
    *baseCmd
}

func new${NameCamel}Cmd() *${Name}Cmd {
	ccmd := &${Name}Cmd{}

	cmd := &cobra.Command{
		Use:   "$Use",
		Short: "$Description",
		Long:  ``$DescriptionLong``,
        Example: ``
$($Examples -join "`n`n")
        ``,
        PreRunE: $PreRunFunction,
		RunE: ccmd.${Name},
    }

    cmd.SilenceUsage = true

    $($CommandArgs.SetFlag -join "`n	")

    // Required flags
    $($CommandArgs.Required -join "`n	")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *${Name}Cmd) ${Name}(cmd *cobra.Command, args []string) error {

    commonOptions, err := getCommonOptions(cmd)
	if err != nil {
        return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

    // query parameters
    queryValue := url.QueryEscape("")
    $RESTQueryBuilder

    // headers
    headers := http.Header{}
    $RestHeaderBuilder

    // form data
    formData := make(map[string]io.Reader)
    $RESTFormDataBuilder

    // body
    body := mapbuilder.NewMapBuilder()
    $RESTBodyBuilder

    // path parameters
    pathParameters := make(map[string]string)
    $RESTPathBuilder
    path := replacePathParameters("${RESTPath}", pathParameters)

    req := c8y.RequestOptions{$RESTHost
        Method:       "${RESTMethod}",
        Path:         path,
        Query:        queryValue,
        Body:         $GetBodyContents,
        FormData:     formData,
        Header:       headers,
        IgnoreAccept: false,
        DryRun:       globalFlagDryRun,
    }

    return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}

"@

    # Must not include BOM!
	$Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
	[System.IO.File]::WriteAllLines($File, $Template, $Utf8NoBomEncoding)

	# Auto format code
	& gofmt -w $File
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

        [string] $Default
    )

    if ($Required -match "true|yes") {
        $Description = "${Description} (required)"
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
    if ($Required -match "true|yes") {
        $Entry | Add-Member -MemberType NoteProperty -Name "Required" -Value "cmd.MarkFlagRequired(`"${Name}`")"
        # $Entry.Required = "cmd.MarkFlagRequired(`"${Name}`")"
    }

    $Entry
}

