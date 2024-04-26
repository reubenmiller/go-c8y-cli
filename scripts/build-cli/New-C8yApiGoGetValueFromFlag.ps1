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

    # TODO: Deprecate .value and use .default (as this is also used by the official go-c8y-cli extensions)
    # $FixedValue = if ($null -ne $Parameters.value) { $Parameters.value } else { $Parameters.default }
    $FixedValue = $Parameters.value

    $FormatValue = ""

    if ($Parameters.format) {
        $FormatValue = ", `"{0}`"" -f $Parameters.format
    }

    $Definitions = @{
        # file (used in multipart/form-data uploads). It writes to the formData object instead of the body
        "file" = "flags.WithFormDataFileAndInfoWithTemplateSupport(cmdutil.NewTemplateResolver(n.factory), `"${prop}`", `"data`"),"

        # fileContents. File contents will be added to body
        "fileContents" = "flags.WithFilePath(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # attachment (used in multipart/form-data uploads), without extra details
        "attachment" = "flags.WithFormDataFile(`"${prop}`", `"data`")...,"

        # multi-part file without extra details and control the form-data field name
        "formDataFile" = "flags.WithFileReader(`"${prop}`", `"${queryParam}`"),"

        # Boolean
        "boolean" = "flags.WithBoolValue(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # Boolean (default, always set the value regardless if )
        "booleanDefault" = "flags.WithDefaultBoolValue(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # Optional fragment (if flag is true)
        "optional_fragment" = "flags.WithOptionalFragment(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # relative datetime
        "datetime" = "flags.WithRelativeTimestamp(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # relative date
        "date" = "flags.WithRelativeDate(false, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # string array/slice
        "string[]" = "flags.WithStringSliceValues(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"

        # string array/slice as a comma separated string
        "stringcsv[]" = "flags.WithStringSliceCSV(`"${prop}`", `"${queryParam}`", `"$FixedValue`"),"
    
        # inventoryChildType
        "inventoryChildType" = "flags.WithInventoryChildType(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # string
        "string" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # stringStatic
        "stringStatic" = "flags.WithStaticStringValue(`"${prop}`", `"$FixedValue`"),"

        # source (special value as powershell need to treat this field as an object)
        "source" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # integer
        "integer" = "flags.WithIntValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # float
        "float" = "flags.WithFloatValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # json_custom: Only supported for use with the body
        "json_custom" = "flags.WithDataValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # binaryUploadURL: uploads a binary and returns the URL 
        "binaryUploadURL" = "c8ybinary.WithBinaryUploadURL(n.factory.Client, n.factory.IOStreams.ProgressIndicator(), `"${prop}`", `"${queryParam}`"$FormatValue),"

        # json - don't do anything because it should be manually set
        "json" = ""

        # tenant
        "tenant" = "flags.WithStringDefaultValue(n.factory.GetTenant(), `"${prop}`", `"${queryParam}`"$FormatValue),"

        # tenantname (optional)
        "tenantname" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # Notifiation2
        "subscriptionName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"
        "subscriptionId" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # application
        "application" = "c8yfetcher.WithApplicationByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        "application_with_versions" = "c8yfetcher.WithApplicationByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # hostedapplication (web app)
        "hostedapplication" = "c8yfetcher.WithHostedApplicationByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # applicationname
        "applicationname" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # microservice
        "microservice" = "c8yfetcher.WithMicroserviceByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # microservice instance
        "microserviceinstance" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # microservice name
        "microservicename" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # uiplugin
        "uiplugin" = "c8yfetcher.WithUIPluginByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # uipluginversion
        "uipluginversion" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # devicerequest array
        "devicerequest[]" = "c8yfetcher.WithIDSlice(args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # id array
        "id[]" = "c8yfetcher.WithIDSlice(args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # software array
        "software[]" = "c8yfetcher.WithSoftwareByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        "softwareDetails" = @(
            "c8yfetcher.WithSoftwareVersionData(n.factory, `"software`", `"version`", `"url`", args, `"`", `"${queryParam}`"$FormatValue),"
        ) -join "`n"

        "configurationDetails" = @(
            "c8yfetcher.WithConfigurationFileData(n.factory, `"configuration`", `"configurationType`", `"url`", args, `"`", `"${queryParam}`"$FormatValue),"
        ) -join "`n"

        # software name
        "softwareName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"
        
        # software version array
        "softwareversion[]" = "c8yfetcher.WithSoftwareVersionByNameFirstMatch(n.factory, `"software`", args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        "deviceservice[]" = "c8yfetcher.WithDeviceServiceByNameFirstMatch(n.factory, `"device`", args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # software version name
        "softwareversionName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # Certificate file
        "certificatefile" = "flags.WithCertificateFile(`"${prop}`", `"${queryParam}`"),"

        # Certificate file
        "certificate[]" = "c8yfetcher.WithCertificateByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"),"

        # firmware array
        "firmware[]" = "c8yfetcher.WithFirmwareByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # firmware name
        "firmwareName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        # firmware version array
        "firmwareversion[]" = "c8yfetcher.WithFirmwareVersionByNameFirstMatch(n.factory, `"firmware`", args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # firmware version name
        "firmwareversionName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"$FormatValue),"

        "firmwareDetails" = @(
            "c8yfetcher.WithFirmwareVersionData(n.factory, `"firmware`", `"version`", `"url`", args, `"`", `"${queryParam}`"),"
        ) -join "`n"

        # firmware version patch array
        "firmwarepatch[]" = "c8yfetcher.WithFirmwarePatchByNameFirstMatch(n.factory, `"firmware`", args, `"${prop}`", `"${queryParam}`"),"

        # firmware patch name
        "firmwarepatchName" = "flags.WithStringValue(`"${prop}`", `"${queryParam}`"),"
        
        # configuration array
        "configuration[]" = "c8yfetcher.WithConfigurationByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"),"
        
        # deviceprofile array
        "deviceprofile[]" = "c8yfetcher.WithDeviceProfileByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"),"
        
        # device array
        "device[]" = "c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # agent array
        "agent[]" = "c8yfetcher.WithAgentByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"


        # devicegroup array
        "devicegroup[]" = "c8yfetcher.WithDeviceGroupByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        
        # smartgroup array
        "smartgroup[]" = "c8yfetcher.WithSmartGroupByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        
        # user array
        "user[]" = "c8yfetcher.WithUserByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"

        # user self url array
        "userself[]" = "c8yfetcher.WithUserSelfByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        
        
        # role self url array
        "roleself[]" = "c8yfetcher.WithRoleSelfByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        
        # role array
        "role[]" = "c8yfetcher.WithRoleByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
        
        # user group array
        "usergroup[]" = "c8yfetcher.WithUserGroupByNameFirstMatch(n.factory, args, `"${prop}`", `"${queryParam}`"$FormatValue),"
    }


    $MatchingType = $Definitions.Keys -eq $Type

    if ($null -eq $MatchingType) {
        # Default to a string
        $MatchingType = "string"
        Write-Warning "Using default type [$MatchingType]"
    }

    # Special type: encoded relative datetime when used as a query parameter
    if ($MatchingType -eq "datetime" -and $SetterType -eq "query") {
        "flags.WithEncodedRelativeTimestamp(`"${prop}`", `"${queryParam}`"$FormatValue),"
    } else {
        $Definitions[$MatchingType]
    }
}
