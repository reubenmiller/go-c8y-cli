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
    
        # string
        "string" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

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
        "application" = "WithApplicationByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"

        # microservice
        "microservice" = "WithMicroserviceByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"

        # device array
        "[]device" = "WithDeviceByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"

        # agent array
        "[]agent" = "WithAgentByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"


        # devicegroup array
        "[]devicegroup" = "WithDeviceGroupByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"
        
        # user array
        "[]user" = "WithUserByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"

        # user self url array
        "[]userself" = "WithUserSelfByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"
        
        
        # role self url array
        "[]roleself" = "WithRoleSelfByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"
        
        # role array
        "[]role" = "WithRoleByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"
        
        # user group array
        "[]usergroup" = "WithUserGroupByNameFirstMatch(args, `"${prop}`", `"${queryParam}`"),"
    }


    $MatchingType = $Definitions.Keys -eq $Type

    if ($null -eq $MatchingType) {
        # Default to a string
        $MatchingType = "string"
        Write-Warning "Using default type [$MatchingType]"
    }

    $Definitions[$MatchingType]
}
