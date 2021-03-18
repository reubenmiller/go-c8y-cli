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

    $Definitions = @{
        # file (used in multipart/form-data uploads). It writes to the formData object instead of the body
        "file" = "flags.WithFormDataFileAndInfo(`"${prop}`", `"data`")...,"

        # fileContents. File contents will be added to body
        "fileContents" = "flags.WithFilePath(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # attachment (used in multipart/form-data uploads), without extra details
        "attachment" = "flags.WithFormDataFile(`"${prop}`", `"data`")...,"

        # Boolean
        "boolean" = "flags.WithBoolValue(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # relative datetime
        "datetime" = "flags.WithRelativeTimestamp(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # string array/slice
        "[]string" = "flags.WithStringSliceValues(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # string array/slice as a comma separated string
        "[]stringcsv" = "flags.WithStringSliceCSV(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"
    
        # string
        "string" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # source (special value as powershell need to treat this field as an object)
        "source" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # integer
        "integer" = "flags.WithIntValue(`"${prop}`", `"${queryParam}`"),"

        # float
        "float" = "flags.WithFloatValue(`"${prop}`", `"${queryParam}`"),"

        # json_custom: Only supported for use with the body
        "json_custom" = "flags.WithDataValue(`"${prop}`", `"${queryParam}`"),"

        # json - don't do anything because it should be manually set
        "json" = ""

        # tenant
        "tenant" = "flags.WithStringDefaultValue(client.TenantName, `"${prop}`", `"${queryParam}`"),"

        # application
        "application" = "c8yfetcher.WithApplicationByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # microservice
        "microservice" = "c8yfetcher.WithMicroserviceByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # device array
        "[]device" = "c8yfetcher.WithDeviceByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # agent array
        "[]agent" = "c8yfetcher.WithAgentByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"


        # devicegroup array
        "[]devicegroup" = "c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # user array
        "[]user" = "c8yfetcher.WithUserByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # user self url array
        "[]userself" = "c8yfetcher.WithUserSelfByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        
        # role self url array
        "[]roleself" = "c8yfetcher.WithRoleSelfByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # role array
        "[]role" = "c8yfetcher.WithRoleByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # user group array
        "[]usergroup" = "c8yfetcher.WithUserGroupByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
    }


    $MatchingType = $Definitions.Keys -eq $Type

    if ($null -eq $MatchingType) {
        # Default to a string
        $MatchingType = "string"
        Write-Warning "Using default type [$MatchingType]"
    }

    $Definitions[$MatchingType]
}
