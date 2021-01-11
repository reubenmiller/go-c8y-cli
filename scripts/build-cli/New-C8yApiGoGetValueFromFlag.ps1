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
        "id" = @{}
        "integer" = @{}
        "json_custom" = @{}
        "microservice" = @{}
        "outputfile" = @{}
        "source" = @{}
        "string" = @{}
        "tenant" = @{}
    }

    # file (used in multipart/form-data uploads). It writes to the formData object instead of the body
    $Setters."file"."query" = "query.Add(`"${queryParam}`", `"true`")"
    $Setters."file"."path" = "pathParameters[`"${queryParam}`"] = `"true`""
    $Setters."file"."body" = "getFileFlag(cmd, `"${prop}`", true, formData)"
    $Definitions."file" = @"
    $($Setters."file".$SetterType)
"@

    # fileContents. File contents will be added to body
    $Setters."fileContents"."body" = "getFileContentsFlag(cmd, `"${prop}`", body)"
    $Definitions."fileContents" = @"
    $($Setters."fileContents".$SetterType)
"@

    # attachment (used in multipart/form-data uploads), without extra details
    $Setters."attachment"."query" = "query.Add(`"${queryParam}`", `"true`")"
    $Setters."attachment"."path" = "pathParameters[`"${queryParam}`"] = `"true`""
    $Setters."attachment"."body" = "getFileFlag(cmd, `"${prop}`", false, formData)"
    $Definitions."attachment" = @"
    $($Setters."attachment".$SetterType)
"@

    # Boolean
    $Setters."boolean"."query" = "query.Add(`"${queryParam}`", fmt.Sprintf(`"%v`", v))"
    $Setters."boolean"."path" = "pathParameters[`"${queryParam}`"] = fmt.Sprintf(`"%v`", v)"
    $Setters."boolean"."body" = "body.Set(`"${queryParam}`", v)"
    if ($FixedValue) {
        $Setters."boolean"."header" = "headers.Add(`"${queryParam}`", `"$FixedValue`")"

        # Note: We have to be carefully when assigning type to the body
        # don't try and convert any types
        if ($FixedValue -is [string]) {
            # Fixed value is a string, so change the value to a string
            $Setters."boolean"."body" = "body.Set(`"${queryParam}`", `"$FixedValue`")"
        } else {
            # -match "^(true|false)$"
            # Fixed value is a boolean, so keep it as a boolean
            $Setters."boolean"."body" = "body.Set(`"${queryParam}`", $FixedValue)"
        }

        $Definitions."boolean" = @"
    if cmd.Flags().Changed("${prop}") {
        if _, err := cmd.Flags().GetBool("${prop}"); err == nil {
            $($Setters."boolean".$SetterType)
        } else {
            return newUserError("Flag does not exist")
        }
    }
"@
    } else {
        $Setters."boolean"."header" = "headers.Add(`"${queryParam}`", fmt.Sprintf(`"%v`", v))"

        $Definitions."boolean" = @"
    if cmd.Flags().Changed("${prop}") {
        if v, err := cmd.Flags().GetBool("${prop}"); err == nil {
            $($Setters."boolean".$SetterType)
        } else {
            return newUserError("Flag does not exist")
        }
    }
"@
    }

    # source
    $Setters."source"."query" = "query.Add(`"${queryParam}`", v)"
    $Setters."source"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."source"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."source" = @"
    if v, err := cmd.Flags().GetString("${prop}"); err == nil {
        $($Setters."source".$SetterType)
    } else {
        return newUserError("Flag does not exist")
    }
"@

    $Setters."datetime"."query" = "query.Add(`"${queryParam}`", v)"
    $Setters."datetime"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."datetime"."body" = "body.Set(`"${queryParam}`", decodeC8yTimestamp(v))"
    $Definitions."datetime" = @"
    if flagVal, err := cmd.Flags().GetString("${prop}"); err == nil && flagVal != "" {
        if v, err := tryGetTimestampFlag(cmd, "${prop}"); err == nil && v != "" {
            $($Setters."datetime".$SetterType)
        } else {
            return newUserError("invalid date format", err)
        }
    }
"@

    # string array
    $Setters."[]string"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(v))"
    $Setters."[]string"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."[]string"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."[]string" = @"
    if items, err := cmd.Flags().GetStringSlice("${prop}"); err == nil {
        if len(items) > 0 {
            for _, v := range items {
                if v != "" {
                    $($Setters."[]string".$SetterType)
                }
            }
        }
    } else {
        return newUserError("Flag does not exist")
    }
"@


    # application
    $Setters."application"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(newIDValue(item).GetID()))"
    $Setters."application"."path" = "pathParameters[`"${queryParam}`"] = newIDValue(item).GetID()"
    $Setters."application"."body" = "body.Set(`"${queryParam}`", newIDValue(item).GetID())"
    $Definitions."application" = @"
    if cmd.Flags().Changed("${prop}") {
        ${prop}InputValues, ${prop}Value, err := getApplicationSlice(cmd, args, "${prop}")

        if err != nil {
            return newUserError("no matching applications found", ${prop}InputValues, err)
        }

        if len(${prop}Value) == 0 {
            return newUserError("no matching applications found", ${prop}InputValues)
        }

        for _, item := range ${prop}Value {
            if item != "" {
                $($Setters."application".$SetterType)
            }
        }
    }
"@

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

    # tenant
    $Setters."tenant"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(v))"
    $Setters."tenant"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."tenant"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."tenant" = @"
    if v := getTenantWithDefaultFlag(cmd, "${prop}", client.TenantName); v != `"`" {
        $($Setters.tenant.$SetterType)
    }
"@

    # string
    $Setters."string"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(v))"
    $Setters."string"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."string"."body" = "body.Set(`"${queryParam}`", v)"
    $Setters."string"."header" = "headers.Add(`"${queryParam}`", v)"
    $Definitions."string" = @"
    if v, err := cmd.Flags().GetString("${prop}"); err == nil {
        if v != "" {
            $($Setters.string.$SetterType)
        }
    } else {
        return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "${prop}", err))
    }
"@

    # outputfile
    <# $Setters."outputfile"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(v))"
    $Setters."outputfile"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."outputfile"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."outputfile" = @"
    if v, err := getOutputFileFlag; err == nil {
        $($Setters.outputfile.$SetterType)
        outputfile = v
    } else {
        return err
    }
"@ #>

    # id
    $Setters."id"."query" = "query.Add(`"${queryParam}`", url.QueryEscape(v))"
    $Setters."id"."path" = "pathParameters[`"${queryParam}`"] = v"
    $Setters."id"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."id" = @"
    if v, err := cmd.Flags().GetString("${prop}"); err == nil {
        if v != "" {
            $($Setters.id.$SetterType)
        }
    } else {
        return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "${prop}", err))
    }
"@

    # integer
    $Setters."integer"."query" = "query.Add(`"${queryParam}`", v)"
    $Setters."integer"."path" = "pathParameters[`"${queryParam}`"] = fmt.Sprintf(`"%d`", v)"
    $Setters."integer"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."integer" = @"
    if v, err := cmd.Flags().GetInt("${prop}"); err == nil {
        $($Setters.integer.$SetterType)
    } else {
        return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "${prop}", err))
    }
"@

    # float
    $Setters."float"."query" = "query.Add(`"${queryParam}`", v)"
    $Setters."float"."path" = "pathParameters[`"${queryParam}`"] = fmt.Sprintf(`"%d`", v)"
    $Setters."float"."body" = "body.Set(`"${queryParam}`", v)"
    $Definitions."float" = @"
    if v, err := cmd.Flags().GetFloat32("${prop}"); err == nil {
        $($Setters.float.$SetterType)
    } else {
        return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "${prop}", err))
    }
"@

    # json_custom: Only supported for use with the body
    $Setters."json_custom"."body" = 'body.Set("{0}", MustParseJSON(v))' -f $queryParam
    $Definitions."json_custom" = @"
    if cmd.Flags().Changed("${prop}") {
        if v, err := cmd.Flags().GetString("${prop}"); err == nil {
            $($Setters.json_custom.$SetterType)
        } else {
            return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "${prop}", err))
        }
    }
"@

    # json - don't do anything because it should be manually set
    $Definitions."json" = ""

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
