# CHANGELOG

## Unreleased

No unreleased features

* Get-Session uses a new c8y session get to retrieve information about the current session
* Fixed bug when using the `-Session` on PUT and POST commands
* Expand-Device supports piping of alarms, events, measurements and operations
* Added `-ProcessingMode` parameter to all commands that use DELETE, PUT and POST requests.

    ```powershell
    New-ManagedObject -Name myobject -ProcessingMode TRANSIENT
    New-ManagedObject -Name myobject -ProcessingMode QUIESCENT
    New-ManagedObject -Name myobject -ProcessingMode PERSISTENT
    New-ManagedObject -Name myobject -ProcessingMode CEP
    ```

## Released

### v1.7.3

* Fixed publishing bug on docker images

### v1.7.2

* Fixed publishing bug on docker images

### v1.7.1

* Fixed publishing bug on docker images
### v1.7.0

* `New-Microservice` requiredRoles are now set when passing the cumulocity.json manifest file to the `-File` parameter
* Added `New-ServiceUser` and `Get-ServiceUser` to create and get a service user that can be used for automation purposes

    ```powershell
    New-ServiceUser -Name "myapp1" -Roles "ROLE_INVENTORY_READ" -Tenants "t12345"

    Get-Serviceuser -Name "myapp1"
    ```
* Fixed target tenant confirmation when using the `-Session` parameter on PUT/POST commands
* `Invoke-ClientRequest`: Added support for `-Template` and `-TemplateVars` parameters
* Removed `-Depth` from internal `ConvertFrom-Json` calls so that the PSc8y is compatible with PowerShell 5.1
* Fixed shallow json conversion bug when using using internal calls to `ConvertFrom-Json` and `ConvertTo-Json`. Max depth of 100 is used on supported PowerShell versions
* `Test-ClientPassphrase` cmdlet to check if passphrase is missing or not. Cmdlet is called automatically when importing the module or calling `set-session`
* `New-User` added support for template and templateVars parameters
* Dry/WhatIf headers are shown in a sorted alphabetically by header name
* Adding Two-Factor-Authentication support
    * TOTP (only)
* Added OAUTH_INTERNAL support
* Encrypting sensitive session information

    ```json
    {
        "credential": {
            "cookies": {
                "0": "{encrypted}abefabefabefabefabefabefabefabefabefabef",
                "1": "{encrypted}abefabefabefabefabefabefabefabefabefabef",
            }
        },
        "password": "{encrypted}abefabefabefabefabefabefabefabefabefabef"
    }
    ```
* Added encrypt/decrypt commands

    ```sh
    $encryptedText=$( c8y sessions encryptText --text "Hello World" )

    c8y sessions decryptText --text "$encryptedText"
    ```

* Fixed broken doc link

### v1.6.0

#### Breaking Changes

* Added support for disabling create/update/delete command individually to prevent accidental data loss. See the [session concept documentation](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/sessions/) for full details.
    * create/update/delete command are disabled by default! They must be enabled otherwise the commands will return an error. Commands can be enabled/disabled from the session properties
    * commands can be temporarily enabled/disabled in the session via environment variables without persisting them in the session settings
    * CI mode which enables all commands via one environment variable

#### Features (PSc8y and c8y)

* Custom rest requests no longer required the `Data` or `File` parameter for `POST` or `PUT` requests. If neither is provided, then the request is sent without a body

    **PowerShell**

    ```powershell
    Invoke-ClientRequest -Uri "/service/exampleMS/myendpoint" -Method "POST"
    ```

    **Bash/zsh**

    ```sh
    c8y rest POST /service/exampleMS/myendpoint
    ```

* Added command to read the current configuration settings as json

    **PowerShell**

    ```powershell
    Get-ClientSetting
    ```

    **Bash/zsh**

    ```sh
    c8y settings list
    ```

* Added support for templates and template variables for all POST and PUT commands. See the [templates concept documentation](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/templates/) for full details.

    `jsonnet` templates can be used to create json data

    *File: custom.device.jsonnet*

    ```jsonnet
    {
        name: "my device",
        type: var("type", "defaultType"),
        cpuThreshold: rand.int,
        c8y_IsDevice: {},
    }
    ```

    **Usage: Bash/zsh**

    ```sh
    c8y inventory create \
        --template ./examples/templates/device.jsonnet \
        --templateVars "type=myCustomType1" \
        --dry
    ```

    **Usage: PowerShell**

    ```sh
    New-ManagedObject `
        -Template ./examples/templates/measurement.jsonnet `
        -TemplateVars "type=myCustomType1" `
        -WhatIf
    ```

    **Output**

    These command would produce the following body which would be sent to Cumulocity.

    ```json
    {
        "name": "my device",
        "type": "myCustomType1",
        "c8y_IsDevice": {},
        "cpuThreshold": 88,
    }
    ```

    To help with the development of templates, there is a command which evaluates a template and prints the output to the console.

    **Bash/zsh**

    ```sh
    c8y template execute --template ./mytemplate.jsonnet
    ```

    **PowerShell**

    ```powershell
    Invoke-Template -Template ./template.jsonnet
    ```

* Added support for setting additional properties when uploading a binary file

    **PowerShell**

    ```powershell
    New-Binary -File "myfile.json" -Data @{ c8y_Global = @{}; type = "c8y_upload" }
    ```

    **Bash/zsh**

    ```sh
    c8y binaries create --file "myfile.json" --data "c8y_Global={},type=c8y_upload"
    ```


* The `Data` parameter now supports a json file path to make it easier to upload complex json structures.

    **Example: Create a new managed object from a json file**

    *./myfile.json*

    ```json
    {
        "name": "server-01",
        "type": "linux",
        "c8y_SoftwareList": [
            { "name": "app1", "version": "1.0.0", "url": ""},
            { "name": "app2", "version": "9", "url": ""},
            { "name": "app3 test", "version": "1.1.1", "url": ""}
        ]
    }
    ```

    Powershell

    ```powershell
    New-ManagedObject -Data ./myfile.json
    ```

    Bash / zsh

    ```sh
    c8y inventory create --data ./myfile.json
    ```

#### PSc8y (PowerShell) minor improvements / fixes

* Fixed logic when removing username information from the current session path when using the hide sensitive information option. Affects MacOS and Linux

* Added Pipeline support to following cmdlets
    * New-TestAlarm
    * New-TestEvent
    * New-TestMeasurement (Confirm impact now set to High)
    * New-TestOperation
    * New-TestUser
    * New-TestGroup
    * New-TestDevice
    * New-TestAgent
    * New-ExternalID

* `New-Microservice`
    * Added `-Key` parameter to allow the user to set a custom value
    * Changed the default value of the `key` property from `{name}-microservice-key` to `{name}` so it matched the default name used by the Cumulocity Java SDK for microservices

### v1.5.2

#### Bugfixes

* Fixed tagging bug which results in the docker images not being tagged with the version number (only the "latest" images were being published)

### v1.5.1

#### Bugfixes

* Fixed bug when using `set-session "my filter"` in bash and zsh profiles which resulted in an "unknown flag" error

### v1.5.0

#### New Features (PSc8y and c8y)

* Added extend paging support and includeAll
    * `includeAll` - Fetch all results. The cli tool will iterate through each of the results and pass them through as they come in
    * `currentPage` - Current page / result set to return
    * `totalPages` - Fetch the given number of pages

    See the [paging docs](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/paging/) for more details and examples.

#### PSc8y (PowerShell)

#### Minor changes

* Renamed `ConvertFrom-Base64ToUtf8` to `ConvertFrom-Base64String`

* Added `ConvertTo-Base64String`

* Renamed `Get-CurrentTenantApplications` to `Get-CurrentTenantApplicationCollection`

* Renamed `Watch-NotificationChannels` to `Watch-NotificationChannel`

* `Watch-*` cmdlets now support piping results as soon as they are received rather than waiting for the duration expire before passing the results back. This enables more complex scenarios, and adhoc event processing tasks

    **Examples**

    Update each alarm which comes in with the serverity CRITICAL. `Update-Alarm` will be run as soon as a result is received, and not just after the 60 second duration of `Watch-Alarm`. 

    ```powershell
    Watch-Alarm -Device 12345 -DurationSec 60 | Update-Alarm -Severity CRITICAL -Force
    ```

    **Complex example**

    Subscribe to realtime alarm notification for a device, and update the alarm severity to CRITICAL if the alarm is active and was first created more than 1 day ago.

    ```powershell
    Watch-Alarm -Device 12345 -DurationSec 600 | Foreach-object {
        $alarm = $_
        $daysOld = ($alarm.time - $alarm.creationTime).TotalDays

        if ($alarm.status -eq "ACTIVE" -and $daysOld -gt 1) {
            $alarm | Update-Alarm -Severity CRITICAL -Force
        }
    }
    ```

* `set-session`: Search now ignores `https://` or `http://` in the url field, as this information is mostly not important when searching for a template. However the full url will still be visible for the user.

#### Bug fixes

* `Get-TenantOptionForCategory`: Removed table view for the tenant option collection output which was causing view problems. Closes #24

    ```powershell
    Get-TenantOptionForCategory -Category application -Verbose

    # outputs
    default.application
    -------------------
    1
    ```

* Fixed parsing of search names with space in their names leading to incorrect application being selected. Closes #22

### v1.4.1

#### Bug fixes

* `set-session`: Fixed bug when setting a session when using the pre-filter which resulted in the wrong session being activated

### v1.4.0

#### PSc8y (PowerShell)

**Renamed cmdlets**

* `New-ChildAssetReference` to `Add-AssetToGroup`
* `Remove-ChildAssetReference` to `Remove-AssetFromGroup`
* `New-ChildDeviceReference` to `Add-ChildDeviceToDevice`
* `Remove-ChildDeviceReference` to `Remove-ChildDeviceFromDevice`

**New cmdlets**

* `Add-DeviceToGroup`
* `Add-ChildGroupToGroup`
* `Remove-DeviceFromGroup`

**Changes**

* `New-TestDevice` has changed to require confirmation before creating the device
* Fixed `Get-Measurement` alias `m` in PSc8y

#### c8y (binary)

* Removed duplicate command `c8y devices find` as query functionality is provided by `c8y devices list --query "name eq 'test*'"`

* Added common options and array items processing to the following commands
    * `c8y agents list`
    * `c8y devices list`
    * `c8y devices listDeviceGroups`

* Added common options to (i.e. --outputFile, --pretty)
    * `c8y micrservices create`
    * `c8y micrservices createHostedApplication`

* Added bash profile script to add support for aliases
* Added guide to creating custom bash aliases
* Removed unreferenced commands

* `c8y session list` Renamed `--filter` to `--sessionFilter` to avoid conflict with the global `--filter` option


#### Docker

* Added docker images to make it easier to try out c8y

    * ghcr.io/reubenmiller/c8y-bash
    * ghcr.io/reubenmiller/c8y-zsh
    * ghcr.io/reubenmiller/c8y-pwsh

#### Docs

* Fixed line wrapping within code blocks. Now horizontal scrollbars are show to preserve the line spacing.
* Added github project link

#### Build

* Improved reliability of realtime api tests
* c8y (golang) binaries are now statically linked (using `CGO_ENABLED=0`) to make them more portable
* Added docker images for pwsh, bash and zsh with c8y already configured

### v1.3.0

**PSc8y (PowerShell)**

* Using a relative date/timestamp of "30" no longer causes a panic when using `DateFrom` and `DateTo` parameters

* Removed unsupported query parameters from Remove-MeasurementCollection
    * `valueFragmentType`
    * `valueFragmentSeries`

* Removed deprecated commands
    `Remove-AuditRecordCollection`

* Fixed `New-Microservice` cmdlet and updated the examples. Resolves #14

* `Set-Session` supports improved filtering
    * Search terms are split on whitespace, and each search term is matched individually against the following session properties
        * index
        * filename (basename only)
        * host
        * tenant
        * username

    * The session can be pre-filtered. Use full when you have a lot of sessions

        ```powershell
        # Powershell
        Set-Session "customer", "dev"

        # Example custom function to set a session for a specific tenant group
        # This can be stored in your powershell profile
        Function Set-CustomerXSession { Set-Session "customer" }
        ```

        ```sh
        # Bash
        c8y session list --filter "customer dev"
        ```

**Build Improvements**

* Updated Pester module to v5
* Updated `go-c8y` dependency to 0.8.0
* Enabled ci test runner
* Added parallel powershell tests execution to reduce test execution times
* Updated golang dependencies

### v1.2.0

**PSc8y (PowerShell)**

* Added Powershell documentation to [online docs](https://reubenmiller.github.io/go-c8y-cli/pwsh/cmdlets/Get-Agent/)

* Removed conflicting cmdlets to make it more obvious which one is correct to use
    * Deleted `Invoke-CustomRestMethod`. Use `Invoke-ClientRequest` instead.
    * Deleted `Invoke-CustomRestRequest`. Use `Invoke-ClientRequest` instead.
    * Moved `Invoke-ClientCommand` to a private command as it is only used internally by other PSc8y cmdlets

* Added additional parameters to `Invoke-ClientRequest`
    * `-Headers`. Add a hashtable of key/values which will be added to headers of the REST request
    * `-ContentType`. Add the `Content-Type` header to the REST request.
    * `-Accept`. Add the `Accept` header to the REST request.
    * `-IgnoreAcceptHeader`. Ignore the accept header when sending the request.
