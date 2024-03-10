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

        [string] $OutputDir = "./",

        [string] $ParentName = ""
    )

    $Name = $Specification.alias.go
	$NameCamel = $Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)

    $FileName = $Specification.alias.go
	$File = Join-Path -Path $OutputDir -ChildPath ("{0}.auto.go" -f $FileName)


    #
    # Meta information
    #
    $Use = $Specification.alias.go
    $Description = $Specification.description
    $DescriptionLong = $Specification.descriptionLong
    $Examples = foreach ($iExample in $Specification.examples.go) {
        if ($iExample.command) {
            $ExampleText = "`$ {0}`n{1}" -f $iExample.command.TrimEnd(), $iExample.description
        } else {
            $ExampleText = $iExample
        }
        $ExampleText
    }
    $RESTPath = $Specification.path -replace " ", "%20"
    $RESTMethod = $Specification.method

    $CommandOptions = New-Object System.Text.StringBuilder

    if ($Specification.hidden) {
        $CommandOptions.AppendLine("`t`tHidden: true,")
    }

    if ($CommandOptions.Length -gt 0) {
        $CommandOptions.Insert(0, "`n")
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

    # Body parameters
    if ($Specification.body) {
        $values = [array]$Specification.body
        $null = $ArgumentSources.AddRange(@($values | ForEach-Object {
            $_ | Add-Member -MemberType NoteProperty -Name "ArgSource" -Value "body"
            $_
        }))
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
    $DeprecationNotice = ""

    if ($Specification.collectionProperty) {
        $collectionProperty = $Specification.collectionProperty
    }

    if ($Specification.deprecated) {
        $DeprecationNotice = $Specification.deprecated
    }

    # Body init
    $RESTBodyBuilderOptions = New-Object System.Text.StringBuilder
    $RESTFormDataBuilderOptions = New-Object System.Text.StringBuilder


    $CompletionBuilderOptions = New-Object System.Text.StringBuilder
    foreach ($iArg in (Remove-SkippedParameters $ArgumentSources)) {
        if ($iArg.pipeline) {
            $PipelineVariableName = $iArg.Name
            $PipelineVariableRequired = if ($iArg.Required) {"true"} else {"false"}
            $PipelineVariableProperty = if ($iArg.Property) { $iArg.Property } else { $iArg.Name }
            $PipelineVariableAliases = $iArg.pipelineAliases
            if (!$PipelineVariableAliases) {
                if ($PipelineVariableName -match "device$" -or $iArg.type -eq "device[]") {
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

            if ($iArg.Type -notmatch "device\b|agent\b|group|devicegroup|self|application|hostedapplication|software\b|softwareName\b|deviceservice\b|softwareversion\b|softwareversionName\b|firmware\b|firmwareName\b|firmwareversion\b|firmwareversionName\b|firmwarepatch\b|firmwarepatchName\b|configuration\b|deviceprofile\b|microservice|id\[\]|devicerequest\[\]") {
                if ($RESTMethod -match "POST" -and $iArg.ArgSource -eq "body") {
                    # Add override capability to piped arguments, so the user can still override piped data with the argument
                    [void] $RESTBodyBuilderOptions.AppendLine("flags.WithOverrideValue(`"$($iarg.Name)`", `"$PipelineVariableProperty`"),")
                }
            }
        }
        if ($iArg.validationSet) {
            $validateSetOptions = @($iArg.validationSet | ForEach-Object { "`"$_`"" }) -join ","
            $null = $CompletionBuilderOptions.AppendLine("completion.WithValidateSet(`"$($iarg.Name)`", $validateSetOptions),")
        }

        # Special system and tenant options completions
        if ($ParentName -match "tenantoptions|systemoptions") {
            $CompletionName = $ParentName + ":" + $iArg.Name
            switch -Regex ($CompletionName) {
                "tenantoptions:category" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithTenantOptionCategory(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
                "tenantoptions:key" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithTenantOptionKey(`"$($iArg.Name)`", `"category`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
                "systemoptions:category" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithSystemOptionCategory(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
                "systemoptions:key" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithSystemOptionKey(`"$($iArg.Name)`", `"category`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
            }
        }

        # Special measurement series/fragments completions
        if ($ParentName -match "measurements") {
            $CompletionName = $ParentName + ":" + $iArg.Name
            switch -Regex ($CompletionName) {
                "measurements:series" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceMeasurementSeries(`"$($iArg.Name)`", `"device`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
                "measurements:valueFragmentType" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceMeasurementValueFragmentType(`"$($iArg.Name)`", `"device`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
                "measurements:valueFragmentSeries" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceMeasurementValueFragmentSeries(`"$($iArg.Name)`", `"device`", `"valueFragmentType`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
            }
        }

        # Special microservices completions
        if ($ParentName -match "loglevels") {
            $CompletionName = $ParentName + ":" + $iArg.Name
            switch -Regex ($CompletionName) {
                "loglevels:loggerName" {
                    [void] $CompletionBuilderOptions.AppendLine("completion.WithMicroserviceLoggers(`"$($iArg.Name)`", `"name`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),")
                }
            }
        }

        # Add Completions based on type
        $ArgType = $iArg.type
        switch ($ArgType) {
            { @("application", "applicationname") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithApplication(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "hostedapplication" { [void] $CompletionBuilderOptions.AppendLine("completion.WithHostedApplication(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "microservice" { [void] $CompletionBuilderOptions.AppendLine("completion.WithMicroservice(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "microservicename" { [void] $CompletionBuilderOptions.AppendLine("completion.WithMicroservice(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "microserviceinstance" { [void] $CompletionBuilderOptions.AppendLine("completion.WithMicroserviceInstance(`"$($iArg.Name)`", `"id`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("role[]", "roleself[]") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithUserRole(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "devicerequest[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceRegistrationRequest(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("user[]", "userself[]") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithUser(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "usergroup[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithUserGroup(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "devicegroup[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceGroup(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "smartgroup[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithSmartGroup(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("tenant", "tenantname") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithTenantID(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "device[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDevice(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "agent[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithAgent(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("software[]", "softwareName") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithSoftware(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("softwareversion[]", "softwareversionName") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithSoftwareVersion(`"$($iArg.Name)`", `"$($iArg.dependsOn | Select-Object -First 1)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("firmware[]", "firmwareName") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithFirmware(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("firmwareversion[]", "firmwareversionName") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithFirmwareVersion(`"$($iArg.Name)`", `"$($iArg.dependsOn | Select-Object -First 1)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            { @("firmwarepatch[]", "firmwarepatchName") -contains $_ } { [void] $CompletionBuilderOptions.AppendLine("completion.WithFirmwarePatch(`"$($iArg.Name)`", `"$($iArg.dependsOn | Select-Object -First 1)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "configuration[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithConfiguration(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "deviceprofile[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceProfile(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "deviceservice[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceService(`"$($iArg.Name)`", `"$($iArg.dependsOn | Select-Object -First 1)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "certificate[]" { [void] $CompletionBuilderOptions.AppendLine("completion.WithDeviceCertificate(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "subscriptionName" { [void] $CompletionBuilderOptions.AppendLine("completion.WithNotification2SubscriptionName(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
            "subscriptionId" { [void] $CompletionBuilderOptions.AppendLine("completion.WithNotification2SubscriptionId(`"$($iArg.Name)`", func() (*c8y.Client, error) { return ccmd.factory.Client()}),") }
        }

        $ArgParams = @{
            Name = $iArg.name
            Type = $iArg.type
            OptionName = $iArg.alias
            Description = $iArg.description
            Default = $iArg.default
            Required = $iArg.required
            Hidden = $iArg.hidden
            Pipeline = $iArg.pipeline
        }
        $CurrentArg = Get-C8yGoArgs @ArgParams
        $null = $CommandArgs.Add($CurrentArg)
    }

    if (!$PipelineVariableName -and $ArgumentSources.Count -gt 0) {
        Write-Warning ("Property is missing pipeline support. cmd={0}" -f @(
            $Specification.name
        ))
    }

    # Post actions
    $PostActionOptions = New-Object System.Text.StringBuilder
    $PostActionsTotal = 0
    foreach ($iArg in (Remove-SkippedParameters $ArgumentSources)) {
        $URLProperty = $iArg.Name
        if ($iArg.Property) {
            $URLProperty = $iArg.Property
        }
        switch ($iArg.Type) {
            "binaryUploadURL" {
                $null = $PostActionOptions.AppendLine("&c8ydata.AddChildAddition{Client: client, URLProperty: `"$URLProperty`"},")
                $PostActionsTotal++
                break
            }
        }
    }
    if ($PostActionsTotal -gt 0) {
        $null = $PostActionOptions.Insert(0, "inputIterators.PipeOptions.PostActions = []flags.Action{`n")
        $null = $PostActionOptions.AppendLine("}")
    }

    # Prepare Request
    $PrepareRequest = New-Object System.Text.StringBuilder

    #
    # Body
    #
    $GetBodyContents = "body"
    $IsFormData = $false
    
    if ($Specification.body) {
        switch ($Specification.bodyContent.type) {
            "binary" {
                $GetBodyContents = "body.GetFileContents()"
                break
            }
            "formdata" {
                $GetBodyContents = "body"
                $IsFormData = $true
            }
            default {
                $GetBodyContents = "body"
                $null = $RESTBodyBuilderOptions.AppendLine("flags.WithDataFlagValue(),")
            }
        }

        $HasProgress = $false
        foreach ($iArg in (Remove-SkippedParameters $Specification.body)) {

            if ($Specification.method -match "POST|PUT" -and -Not $HasProgress) {
                if ($iArg.type -in @("file", "fileContents", "attachment")) {
                    $HasProgress = $true
                    $null = $PrepareRequest.Append("PrepareRequest: c8ybinary.AddProgress(cmd, `"$($iArg.name)`", cfg.GetProgressBar(n.factory.IOStreams.ErrOut, n.factory.IOStreams.IsStderrTTY())),")
                }
            }

            if ($iArg.options -and $iArg.options.formData) {
                $null = $RESTFormDataBuilderOptions.AppendLine("Append(flags.WithFormDataProperty(`"{0}`"))." -f $iArg.options.formData)
            }

            $code = New-C8yApiGoGetValueFromFlag -Parameters $iArg -SetterType "body"
            if ($code) {
                switch -Regex ($code) {
                    "^flags\.WithFormData" {
                        # $null = $RESTFormDataBuilderOptions.AppendLine($code)
                        $null = $RESTFormDataBuilderOptions.AppendLine("Append({0}...)." -f $code.TrimEnd(',').TrimEnd('.'))
                        break
                    }

                    "^(flags\.|c8yfetcher\.|With|c8ybinary\.)" {
                        if ($IsFormData) {
                            $null = $RESTFormDataBuilderOptions.AppendLine("Append({0})." -f $code.TrimEnd(','))
                        } else {
                            $null = $RESTBodyBuilderOptions.AppendLine($code)
                        }
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
                SetFlagOptions = @(
                    "f.WithTemplateFlag(cmd)"
                )
            }
        }

        #
        # Apply a body template to the data
        #
        if ($Specification.bodyTemplates) {    
            foreach ($BodyTemplate in $Specification.bodyTemplates) {
                switch ($BodyTemplate.type) {
                    "jsonnet" {
                        # ApplyLast: true == apply template to the existing json (potentially overriding values)
                        #            false == Use template as base json, and the existing json will take precendence
                        if ($BodyTemplate.applyLast -eq "true") {
                            $null = $RESTBodyBuilderOptions.AppendLine("flags.WithRequiredTemplateString(```n{0}``)," -f @(
                                $BodyTemplate.template
                            ))
                            
                        } else {
                            $null = $RESTBodyBuilderOptions.AppendLine("flags.WithDefaultTemplateString(```n{0}``)," -f @(
                                $BodyTemplate.template
                            ))
                        }
    
                    }
                    "none" {
                        # Do nothing
                    }
                    default {
                        Write-Warning ("Unsupported templating type [{0}]" -f $BodyTemplate.type)
                    }
                }
            }
        }

        #
        # Add support for user defined templates to control body
        #
        if ($Specification.bodyTemplates.type -ne "none") {
            if ($IsFormData) {
                $null = $RESTFormDataBuilderOptions.AppendLine("Append({0})." -f "cmdutil.WithTemplateValue(n.factory)")
                $null = $RESTFormDataBuilderOptions.AppendLine("Append({0})." -f "flags.WithTemplateVariablesValue()")
            } else {
                $null = $RESTBodyBuilderOptions.AppendLine("cmdutil.WithTemplateValue(n.factory),")
                $null = $RESTBodyBuilderOptions.AppendLine("flags.WithTemplateVariablesValue(),")
            }
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
    $CumulocityQueryExpressionBuilder = New-Object System.Text.StringBuilder
    $RESTQueryBuilderPost = New-Object System.Text.StringBuilder

    if ($Specification.queryParameters) {
        # TODO: Handle special case of an item with .children
        foreach ($Properties in (Remove-SkippedParameters $Specification.queryParameters)) {

            if ($Properties.type -eq "queryExpression" -and $Properties.children) {

                $null = $CumulocityQueryExpressionBuilder.AppendLine("		flags.WithCumulocityQuery(")
                $null = $CumulocityQueryExpressionBuilder.AppendLine("			[]flags.GetOption{")
                
                foreach ($child in $Properties.children) {

                    # Ignore special in-built values as these are handled separately
                    if ($child.name -in @("queryTemplate", "orderBy")) {
                        continue
                    }

                    # Special case to handle Cumulocity query language builder
                    $code = New-C8yApiGoGetValueFromFlag -Parameters $child -SetterType "query"
                    if ($code) {
                        $null = $CumulocityQueryExpressionBuilder.AppendLine($code)
                    }
                }
                $null = $CumulocityQueryExpressionBuilder.AppendLine("			},")
                $null = $CumulocityQueryExpressionBuilder.AppendLine("			`"$($Properties.property || $Properties.name)`",")
                $null = $CumulocityQueryExpressionBuilder.AppendLine("		),")

            } else {
                $code = New-C8yApiGoGetValueFromFlag -Parameters $Properties -SetterType "query"
                if ($code) {
                    $null = $RESTQueryBuilderWithValues.AppendLine($code)
                }
            }
        }
    }
    if ($Specification.method -match "GET") {
        
        $null = $RESTQueryBuilderPost.AppendLine(@"
        commonOptions, err := cfg.GetOutputCommonOptions(cmd)
        if err != nil {
            return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
        }
"@)
        # Support mapping common flags to custom query parameters (e.g. if a service uses limit instead of pageSize)
        # but for consistency the flag --pageSize will still be used and mapped to 'limit'.
        if ($Specification.flagMapping) {
            $flagAliases = foreach ($item in $Specification.flagMapping.PSObject.Properties) {
                "`"$($item.Name)`":`"$($item.Value)`""
            }
            $null = $RESTQueryBuilderPost.AppendLine("commonOptions.AddQueryParametersWithMapping(query, map[string]string{$($flagAliases -join ',')})")
        } else {
            $null = $RESTQueryBuilderPost.AppendLine("commonOptions.AddQueryParameters(query)")
        }
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

    if ($Specification.contentType) {
        $null = $RestHeaderBuilderOptions.AppendLine("flags.WithStaticStringValue(`"Content-Type`", `"$($Specification.contentType)`"),")
    }

    if ($Specification.addAccept -and $Specification.accept) {
        $null = $RestHeaderBuilderOptions.AppendLine("flags.WithStaticStringValue(`"Accept`", `"$($Specification.accept)`"),")
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

    # Add common parameters
    $FlagBuilderOptions = New-Object System.Text.StringBuilder
    if ($Specification.method -match "DELETE|PUT|POST") {
        $null = $FlagBuilderOptions.AppendLine("flags.WithProcessingMode(),")
    }


    #
    # Pre run validation (disable some commands without switch flags)
    #
     $FunctionalMethod = if ($Specification.semanticMethod) {
        $Specification.semanticMethod
    } else {
        $Specification.method
    }

    $PreRunFunction = switch ($FunctionalMethod) {
        "POST" { "f.CreateModeEnabled()" }
        "PUT" { "f.UpdateModeEnabled()" }
        "DELETE" { "f.DeleteModeEnabled()" }
        default { "nil" }
    }

    # Additional options
    $RequestOptionsBuilder = New-Object System.Text.StringBuilder
    if ($Specification.responseType -eq "array") {
        $null = $RequestOptionsBuilder.AppendLine("ResponseData: make([]map[string]interface{}, 0),")
    } elseif ($Specification.responseType -eq "object") {
        # Do nothing, so it already defaults to a map[string]interface{}
    }

    #
    # Template
    #
    $Template = @"
// Code generated from specification version 1.0.0: DO NOT EDIT
package $($Name.ToLower())

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

    "github.com/MakeNowJust/heredoc/v2"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
    "github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// ${NameCamel}Cmd command
type ${NameCamel}Cmd struct {
    *subcommand.SubCommand

    factory *cmdutil.Factory
}

// New${NameCamel}Cmd creates a command to $Description
func New${NameCamel}Cmd(f *cmdutil.Factory) *${NameCamel}Cmd {
	ccmd := &${NameCamel}Cmd{
        factory: f,
    }
	cmd := &cobra.Command{
		Use:   "$Use",
		Short: "$Description",
		Long:  ``$DescriptionLong``,$CommandOptions
        Example: heredoc.Doc(``
$($Examples -join "`n`n")
        ``),
        PreRunE: func(cmd *cobra.Command, args []string) error {
            return $PreRunFunction
        },
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
            if ($FlagBuilderOptions) {
                $FlagBuilderOptions.ToString().TrimEnd()
            }
        )
        $(
            if ($CommandArgs.SetFlagOptions) {
                ($CommandArgs.SetFlagOptions -join ",`n") + ","
            }
        )
        $(
            if ($PipelineVariableAliases) {
                $aliases = ($PipelineVariableAliases | ForEach-Object { "`"$_`""` }) -join ", "
                "flags.WithExtendedPipelineSupport(`"$PipelineVariableName`", `"$PipelineVariableProperty`", $PipelineVariableRequired, $aliases),"
            } else {
                "flags.WithExtendedPipelineSupport(`"$PipelineVariableName`", `"$PipelineVariableProperty`", $PipelineVariableRequired),"
            }   
        )
        $(
            foreach ($item in $CommandArgs) {
                $iAliases = $item.PipelineAliases
                if (!$iAliases) {
                    $iAliases = @($item.Name)
                }
                if ($item.PipelineAliases) {
                    $usedAliases = @{}
                    $sourceAliases = ($item.PipelineAliases | ForEach-Object {
                        if (!$usedAliases.ContainsKey($_)) {
                            "`"$_`""`
                        }
                        $usedAliases[$_] = $true
                    } | Where-Object { $_ }) -join ", "
                    if ($sourceAliases) {
                        "flags.WithPipelineAliases(`"$($item.Name)`", $sourceAliases),`n"
                    }
                }
            }
        )
        $(
            if ($collectionProperty) {
                "flags.WithCollectionProperty(`"$collectionProperty`"),`n"
            }
            if ($DeprecationNotice) {
                "flags.WithDeprecationNotice(`"$DeprecationNotice`"),`n"
            }
            if ($Specification.semanticMethod) {
                "flags.WithSemanticMethod(`"$($Specification.semanticMethod)`"),`n"
            }
        )
	)

    // Required flags
    $($CommandArgs.Required -join "`n	")
    $($CommandArgs.Hidden -join "`n	")

    ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *${NameCamel}Cmd) RunE(cmd *cobra.Command, args []string) error {
    cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
    // Runtime flag options
    flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
    client, err := n.factory.Client()
	if err != nil {
		return err
	}
    inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
    if err != nil {
        return err
    }

    // query parameters
    query := flags.NewQueryTemplate()
    err = flags.WithQueryParameters(
		cmd,
        query,
        inputIterators,
        flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
        $RESTQueryBuilderWithValues
        $CumulocityQueryExpressionBuilder
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
        flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
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
        inputIterators,$(
            if ($RESTFormDataBuilderOptions.Length -gt 0) {
                @"
                flags.WithOptionBuilder().
                    $RESTFormDataBuilderOptions
                Build()...
"@
            }
        )
    )
    if err != nil {
		return cmderrors.NewUserError(err)
    }
    

    // body
    body := mapbuilder.NewInitializedMapBuilder($(($RESTMethod -match "PUT|POST" -and $Specification.body).ToString().ToLower()))
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
        IgnoreAccept: cfg.IgnoreAcceptHeader(),
        DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
        $PrepareRequest
        $RequestOptionsBuilder
    }
    $PostActionOptions

    return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
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

    [array]($CommandParameters | Where-Object {
        if ($_.skip -eq $true) {
            Write-Verbose ("Skipping parameter [{0}] as it is marked as skip" -f $_.name)
        }
        $_.skip -ne $true
    })
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

        [string] $Hidden,

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
        "json" {
            @{
                SetFlagOptions = @(
                    "flags.WithData()"
                    "f.WithTemplateFlag(cmd)"
                ) 
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

        { @("datefrom", "dateto", "datetime", "date") -contains $_ } {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
                PipelineAliases = @("time", "creationTime", "creationTime", "lastUpdated")
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
                PipelineAliases = @("id", "source.id", "managedObject.id", "deviceId")
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

        "string[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSlice(`"${Name}`", `"${OptionName}`", []string{`"${Default}`"}, `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "stringcsv[]" {
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

        "device[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("deviceId", "source.id", "managedObject.id", "id")
            }
        }

        "agent[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("deviceId", "source.id", "managedObject.id", "id")
            }
        }

        # Management repository types
        { $_ -in "software[]", "softwareversion[]", "firmware[]", "firmwareversion[]", "firmwarepatch[]", "configuration[]", "deviceprofile[]" } {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        # Device extensions
        { $_ -in @("deviceservice[]") } {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        # Management name lookup
        { $_ -in "softwareName", "softwareversionName", "firmwareName", "firmwareversionName", "firmwarepatchName" } {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringP(`"${Name}`", `"${Default}`", `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().String(`"${Name}`", `"${Default}`", `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "binaryUploadURL" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "devicegroup[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("source.id", "managedObject.id", "id")
            }
        }

        "id[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "smartgroup[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("managedObject.id")
            }
        }

        "roleself[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("self", "id")
            }
        }

        "role[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("id")
            }
        }

        "usergroup[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("id")
            }
        }

        {$_ -in "application", "applicationname", "hostedapplication"} {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
                PipelineAliases = @("id")
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
                PipelineAliases = @("id")
            }
        }

        "microserviceinstance" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
                PipelineAliases = @("id")
            }
        }

        "microservicename" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }
            @{
                SetFlag = $SetFlag
                PipelineAliases = @("name")
            }
        }

        "inventoryChildType" {
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

        { $_ -in "subscriptionName", "subscriptionId" } {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "stringStatic" {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "devicerequest[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
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

        {$_ -in "tenant", "tenantname"}  {
            $SetFlag = if ($UseOption) {
                'cmd.Flags().StringP("{0}", "{1}", "{2}", "{3}")' -f $Name, $OptionName, $Default, $Description
            } else {
                'cmd.Flags().String("{0}", "{1}", "{2}")' -f $Name, $Default, $Description
            }

            @{
                SetFlag = $SetFlag
                PipelineAliases = @("tenant", "owner.tenant.id")
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

        "formDataFile" {
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

        {$_ -in "boolean", "booleanDefault", "optional_fragment"} {
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

        "user[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        "userself[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSliceP(`"${Name}`", []string{`"${Default}`"}, `"${OptionName}`", `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }

            @{
                SetFlag = $SetFlag
            }
        }

        # Trusted device certficates
        "certificate[]" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().StringSlice(`"${Name}`", `"${OptionName}`", []string{`"${Default}`"}, `"${Description}`")"
            } else {
                "cmd.Flags().StringSlice(`"${Name}`", []string{`"${Default}`"}, `"${Description}`")"
            }
            @{
                SetFlag = $SetFlag
            }
        }

        "certificatefile" {
            $SetFlag = if ($UseOption) {
                "cmd.Flags().String(`"${Name}`", `"${OptionName}`", `"${Default}`", `"${Description}`")"
            } else {
                "cmd.Flags().String(`"${Name}`", `"${Default}`", `"${Description}`")"
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

    if ($Hidden -match "true|yes" -and $Pipeline -notmatch "true") {
        $Entry | Add-Member -MemberType NoteProperty -Name "Hidden" -Value "_ = cmd.Flags().MarkHidden(`"${Name}`")"
    }

    $Entry.Name = $Name

    $Entry
}

