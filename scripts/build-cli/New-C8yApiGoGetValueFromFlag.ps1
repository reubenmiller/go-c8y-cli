Function New-C8yApiGoGetValueFromFlag {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true
        )]
        [object] $Parameters,

        [Parameter(
            Mandatory = $true
        )]
        [ValidateSet("query", "path", "body", "header")]
        [string] $SetterType
    )
 
    $prop = $Parameters.name
    $queryParam = $Parameters.property
    if (!$queryParam) {
        $queryParam = $Parameters.name
    }

    $Type = $Parameters.type

    $FixedValue = $Parameters.value

    $Definitions = @{}
    $Setters = @{
        "[]agent" = @{}
        "[]device" = @{}
        "[]devicegroup" = @{}
        "[]role" = @{}
        "[]roleself" = @{}
        "[]string" = @{}
        "[]user" = @{}
        "[]usergroup" = @{}
        "[]userself" = @{}
        "application" = @{}
        "boolean" = @{}
        "datetime" = @{}
        "float" = @{}
        "file" = @{}
        "fileContents" = @{}
        "attachment" = @{}
        "integer" = @{}
        "json_custom" = @{}
        "microservice" = @{}
        "string" = @{}
        "tenant" = @{}
    }

    # file (used in multipart/form-data uploads). It writes to the formData object instead of the body
    $Definitions."file" = "flags.WithFormDataFileAndInfo(`"${prop}`", `"data`")...,"

    # fileContents. File contents will be added to body
    $Definitions."fileContents" = "flags.WithFilePath(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

    # attachment (used in multipart/form-data uploads), without extra details
    $Definitions."attachment" = "flags.WithFormDataFile(`"${prop}`", `"data`")...,"

    # Boolean
    $Definitions."boolean" = "flags.WithBoolValue(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

    # relative datetime
    $Definitions."datetime" = "flags.WithRelativeTimestamp(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

    # string array/slice
    $Definitions."[]string" = "flags.WithStringSliceValues(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"
   
    # string
    $Definitions."string" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

    # integer
    $Definitions."integer" = "flags.WithIntValue(`"${prop}`", `"${queryParam}`"),"

    # float
    $Definitions."float" = "flags.WithFloatValue(`"${prop}`", `"${queryParam}`"),"

    # json_custom: Only supported for use with the body
    $Definitions."json_custom" = "flags.WithDataValue(`"${prop}`", `"${queryParam}`"),"

    # json - don't do anything because it should be manually set
    $Definitions."json" = ""

    # tenant
    $Definitions."tenant" = "flags.WithStringDefaultValue(client.TenantName, `"${prop}`", `"${queryParam}`"),"

    #
    # Old style getters
    #

    # application
    $Definitions."application" = "WithApplicationReferenceByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"

    # microservice
    $Setters."microservice"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(newIDValue(item).GetID()))"
    $Setters."microservice"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."microservice"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."microservice" = @"
    if cmd.Flags().Lookup("${prop}") != nil {
        ${prop}InputValues, ${prop}Value, err := getMicroserviceSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching microservices found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching microservices found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."microservice".$SetterType)
            }
        }
    }
"@

    # device array
    $Setters."[]device"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]device"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]device"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]device" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedDeviceSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching devices found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching devices found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]device".$SetterType)
            }
        }
    }
"@

    # agent array
    $Setters."[]agent"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]agent"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]agent"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]agent" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedAgentSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching agents found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching agents found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]agent".$SetterType)
            }
        }
    }
"@

    # devicegroup array
    $Setters."[]devicegroup"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]devicegroup"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]devicegroup"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]devicegroup" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedDeviceGroupSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching device groups found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching device groups found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]devicegroup".$SetterType)
            }
        }
    }
"@

    # user array
    $Setters."[]user"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]user"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]user"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]user" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedUserSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching users found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching users found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]user".$SetterType)
            }
        }
    }
"@

    # user self url array
    $Setters."[]userself"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]userself"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]userself"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]userself" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedUserLinkSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching users found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching users found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]userself".$SetterType)
            }
        }
    }
"@

    # role self url array
    $Setters."[]roleself"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]roleself"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]roleself"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]roleself" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedRoleSelfSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching roles found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching roles found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]roleself".$SetterType)
            }
        }
    }
"@
    # role array
    $Setters."[]role"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]role"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]role"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]role" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedRoleSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching roles found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching roles found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]role".$SetterType)
            }
        }
    }
"@

    # user group array
    $Setters."[]usergroup"."query" = "query.Add(`"${queryParam}`", newIDValue(item).GetID())"
    $Setters."[]usergroup"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."[]usergroup"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."[]usergroup" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getFormattedGroupSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching user groups found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching user groups found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."[]usergroup".$SetterType)
            }
        }
    }
"@


    $MatchingType = $Definitions.Keys -eq $Type

    if ($null -eq $MatchingType) {
        # Default to a string
        $MatchingType = "string"
        Write-Warning "Using default type [$MatchingType]"
    }

    $Definitions[$MatchingType]
}

#
# Definitions
#
