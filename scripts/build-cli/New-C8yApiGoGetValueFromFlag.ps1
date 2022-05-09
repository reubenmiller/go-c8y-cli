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

        # Boolean (default, always set the value regardless if )
        "booleanDefault" = "flags.WithDefaultBoolValue(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # Optional fragment (if flag is true)
        "optional_fragment" = "flags.WithOptionalFragment(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # relative datetime
        "datetime" = "flags.WithRelativeTimestamp(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # relative date
        "date" = "flags.WithRelativeDate(false, `"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

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

        # binaryUploadURL: uploads a binary and returns the URL 
        "binaryUploadURL" = "flags.WithBinaryUploadURL(client, `"${prop}`", `"${queryParam}`"),"

        # json - don't do anything because it should be manually set
        "json" = ""

        # tenant
        "tenant" = "flags.WithStringDefaultValue(client.TenantName, `"${prop}`", `"${queryParam}`"),"

        # tenantname (optional)
        "tenantname" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # application
        "application" = "c8yfetcher.WithApplicationByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # hostedapplication (web app)
        "hostedapplication" = "c8yfetcher.WithHostedApplicationByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # applicationname
        "applicationname" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # microservice
        "microservice" = "c8yfetcher.WithMicroserviceByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # microservice instance
        "microserviceinstance" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # devicerequest array
        "[]devicerequest" = "c8yfetcher.WithIDSlice(args, `"${prop}`", `"${queryParam}`"),"

        # id array
        "[]id" = "c8yfetcher.WithIDSlice(args, `"${prop}`", `"${queryParam}`"),"

        # software array
        "[]software" = "c8yfetcher.WithSoftwareByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        "softwareDetails" = @(
            "c8yfetcher.WithSoftwareVersionData(client, `"software`", `"version`", `"url`", args, `"`", `"${queryParam}`"),"
        ) -join "`n"

        # software name
        "softwareName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"
        
        # software version array
        "[]softwareversion" = "c8yfetcher.WithSoftwareVersionByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # software version name
        "softwareversionName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # firmware array
        "[]firmware" = "c8yfetcher.WithFirmwareByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # firmware name
        "firmwareName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        # firmware version array
        "[]firmwareversion" = "c8yfetcher.WithFirmwareVersionByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # firmware version name
        "firmwareversionName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"

        "firmwareDetails" = @(
            "c8yfetcher.WithFirmwareVersionData(client, `"firmware`", `"version`", `"url`", args, `"`", `"${queryParam}`"),"
        ) -join "`n"

        # firmware version patch array
        "[]firmwarepatch" = "c8yfetcher.WithFirmwarePatchByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # firmware patch name
        "firmwarepatchName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"
        
        # configuration array
        "[]configuration" = "c8yfetcher.WithConfigurationByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # deviceprofile array
        "[]deviceprofile" = "c8yfetcher.WithDeviceProfileByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # device array
        "[]device" = "c8yfetcher.WithDeviceByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"

        # agent array
        "[]agent" = "c8yfetcher.WithAgentByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"


        # devicegroup array
        "[]devicegroup" = "c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
        # smartgroup array
        "[]smartgroup" = "c8yfetcher.WithSmartGroupByNameFirstMatch(client, args, `"${prop}`", `"${queryParam}`"),"
        
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

    # Special type: encoded relative datetime when used as a query parameter
    if ($MatchingType -eq "datetime" -and $SetterType -eq "query") {
        "flags.WithEncodedRelativeTimestamp(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"
    } else {
        $Definitions[$MatchingType]
    }
}
