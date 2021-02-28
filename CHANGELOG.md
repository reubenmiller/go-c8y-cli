# CHANGELOG

## Known limitations

Usage of large numbers (using exponential notation) in templates can cause the server to respond with a 422 status code. This is due to a current limitation in the jsonnet template library.

As a workaround, use the `data` parameter when working with large numbers (and don't specify a `template` parameter). This will preserve the number format as provided.

**Example**

*datapoint.largeint.json*

```json
{
    "c8y_Kpi": {
        "max": 19.1010101E19,
        "description": "Counter"
    }
}
```

```sh
c8y inventory create --data "./datapoint.largeint.json"
```

## Failing tests

None :)

## TODO

* Display session information after selecting a session (like in powershell but natively)
* Option to not store the password in the session. It should force the user to use their password each time (then delete the password value?)
* Convert hashtable to json, i.e. or reject it automatically? replace ";" with ",", and replace "@{" with "{", "=" with ":"

* Add confirmation prompts to c8y binary
    ```json
    {
        "c8y_SoftwareList": [
            {"name": "test1", "version": "1.0.0"}
        ],
        "name": "myname"
    }
    ```


    ```js
    name
    c8y_SoftwareList.0.name
    c8y_SoftwareList.0.version
    c8y_SoftwareList.1.name
    c8y_SoftwareList.1.version
    ```


    ```sh
    *.*name
    *.*name

    *.#.*name
    ```

  * Create activity log cmdlets?
    * Get recent history, get commands with filter (api, host, method etc.)


## Unreleased

No unreleased features

* Added links to native go command in PowerShell help in the `.LINK` (related links) section
* `Get-TenantVersion` returns an object instead of a single value.
    ```json
    {
        "category": "system",
        "key": "version",
        "value": "1006.6.0"
    }
    ```
* Session folder is created automatically if it does not yet exist when creating a new session
* --outputFile appends a new line character to the output when the Content-Type header contains "json"

* Removed `TimeoutSec` to `Timeout`
* `--timeout` has been changed to accept seconds not milliseconds. However you can enter "0.001" if you really want millisecond accuracy (not sure why though)
* Added support for flag value completions (i.e. )
    
    ```sh
    c8y operations list --status <TAB><TAB>
    ```

* Adding short options for various global options
  * `-p` => `--pageSize`
  * `-t` => `--withTotalPages`
  * `-r` => `--raw`

* Added colorize and pretty print options to output formatting
    * pretty print when using streaming (`-c` or `--compress`)
    * colorize json and json line output (`--noColor` or `-M` (for monochromatic output)). Output is colorized by default when writing to Terminal, but it can be disabled using `--noColor` or `-M`
* `--select` parameter now supports advanced property selection using dot notation, wildcards (including globstar)
* Added new `--noAccept` global parameter to ignore the Accept header. This usually only affects `PUT` and `POST` requests.

* Fixed #43. New-ServiceUser now accepts more than 1 role
* Added Expand-DeviceGroup cmdlet
* Adding following template variables
    * time.now
    * time.nowNano
    * input.index (current iteration when using pipelines)

* TODO: Create custom ConvertFrom-Json
    * Replace ConvertFrom-Json -Depth calls in code as the default is already high enough (1024). This simpifies the code as the check for powershell version can be ignored
    * add options to strip out cumulocity noise (i.e. additionParents etc.)
* TODO: Make -InformationVariable or at least ErrorVariable work when using IncludeAll

* Optimized inventory queries when using the `includeAll` parameter on inventory managed objects
    ```sh
    $filter=_id gt '{lastId}' $orderby=_id asc
    ```


* New cmdlet `ConvertTo-NestedJson` which is a proxy function of `ConvertTo-Json` where depth is set to 10 by default
* Removed `--format` as `select` and `csv` can be used for the same effect (and more).

    ```
    id=$( c8y devices list --select id --csv | head -1 )
    ```

### PSc8y

* New alias to the c8y binary. This makes it easier to run the binary directly without the powershell wrapper.

    When using the c8y binary directly, the parameter names are slightly different, and should start with "--" and not the normalized powershell "-".

    ```powershell
    c8y devices list --pageSize 1

    c8y help
    ```

* [Get-MeasurementCollection / c8y measurements list]
    * Renamed `csv` parameter to `csvFormat` to avoid conflict with new global csv (client side) parameter
    * Renamed `excel` parameter to `excelFormat` to be consistent with the renamed `csvFormat` parameter

* Added support for passing non-pipeline parameters via the object. The id field will be read from the object. #42
    ```powershell
    $Group = New-TestDeviceGroup
    $Device = New-TestDevice
    Add-DeviceToGroup -Group $Group -NewChildDevice $Device
    ```

### Breaking changes

#### PowerShell

* Renamed all user group related cmdelts from "Group" to "UserGroup"
    |From|New Name|
    |-|-|
    |New-Group|New-UserGroup|
    |Get-Group|Get-UserGroup|
    |Get-GroupByName|Get-UserGroupByName|
    |Get-GroupCollection|Get-UserGroupCollection|
    |Get-GroupMembershipCollection|Get-UserGroupMembershipCollection|
    |New-TestGroup|New-TestUserGroup|
    |Update-Group|Update-UserGroup|
    |Remove-Group|Remove-UserGroup|

#### Shell

* c8y rest `--ignoreAcceptHeader` has been removed as there is a global `--noAccept` header

* `realtime subscribe`: Remove shorthand `-c` variant of `--channel` as it conflicts with the new `--compress`, `-c` global flag.
* `Get-ManagedObjectCollect` renamed `Device` parameter to `Ids` and removed device lookup as the parameter is related to a generic managed object and not a device.
* Renamed json output parameters
    * `--pretty` has been renamed to `--compress` (or `-c`) to be inline with popular json utility jq.
    * Invoke-ClientRequest has both `-Pretty` and `-Compress` options - renamed `Pretty` to `Compress`

### New Features

#### Activity log
* Records c8y commands (c8y binary arguments only) and HTTP request/response meta information
* Enable the logs for specific sessions or have a global setting
* Disable activity log for single commands by using `noLog` parameter

    ```sh
    c8y devices create --name "myDeviceName" --noLog
    ```

**Example**

*File: c8y.activitylog.2021.02.11.json*

```json
{"time":"2021-02-11T07:07:29.7405634Z","ctx":"gOBHCQDe","type":"command","arguments":["devices","create","--name=myDevice01"]}
{"time":"2021-02-11T07:07:29.9139964Z","ctx":"gOBHCQDe","type":"request","method":"POST","host":"c8y.example.com","path":"/inventory/managedObjects","query":"","accept":"application/json","processingMode":"","statusCode":201,"responseTimeMS":125,"responseSelf":"https://c8y.example.com/inventory/managedObjects/367930"}
{"time":"2021-02-11T07:08:10.8468454Z","ctx":"gCVnTfZt","type":"command","arguments":["devices","update","--id","367930","--newName","my new device"]}
{"time":"2021-02-11T07:08:11.0067532Z","ctx":"gCVnTfZt","type":"request","method":"PUT","host":"c8y.example.com","path":"/inventory/managedObjects/367930","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":146,"responseSelf":"https://c8y.example.com/inventory/managedObjects/367930"}
{"time":"2021-02-11T07:08:26.9742641Z","ctx":"qFRWdUsT","type":"command","arguments":["devices","delete","--id=367930"]}
{"time":"2021-02-11T07:08:27.1066364Z","ctx":"qFRWdUsT","type":"request","method":"DELETE","host":"c8y.example.com","path":"/inventory/managedObjects/367930","query":"","accept":"application/json","processingMode":"","statusCode":204,"responseTimeMS":122,"responseSelf":""}
```

**Settings**

The activity log settings can be set for individual c8y sessions or globally in your `settings.json` file.

```json
{
    "settings": {
        "activityLog": {
            "enabled": true,
            // defaults to current working directory if not specified
            "path": "/home/myuser/.c8y-activitylogs/",
            "methodFilter": "GET PUT POST DELETE"
        }
    }
}
```

### Bash/zsh improvements

**Pipeline support**

* Added pipelines to most commands

    ```sh
    c8y alarms --device 12345 --status ACTIVE | c8y alarms update --status CLEARED
    ```

    ```sh
    c8y devices list | c8y operations list
    ```

* When doing batch size of 1, only set the exit code to the last value
* Maximum jobs limit to protect against unexpected updates
* New global parameters
    * `--compact` or `-c` compact the json removing any indentation
    * `--stream` will automatically convert a json array to individual json lines (objects)
        ```json
        [
            { "id": "1"},
            { "id": "2"}
        ]
        ```

        Will be transformed individually stream json objects. Each object will be written to stdout as it is processed.

        ```json
        { "id": "1"}
        { "id": "2"}
        ```

**Templates**

* When using piped data the data is stored for use in a template variable `input.value`.

    **Example: Add a fragment based on the type name**

    ```sh
    seq 1 | c8y devices create --data "type=myvalue" \
        | c8y devices update --template "{ [input.value.type]: {} }"
    ```

    **Example: Add a type to devices which have a specific fragment and also mark when it was updated**

    ```sh
    c8y devices list --query "has(c8y_myCustomType)" \
        | c8y devices update --template "{ type: 'c8y_Linux', c8y_jobInfo: { job: input.index, lastUpdated: time.now } }"
    ```

**Progress bar (alpha)**

* `--progress` bar can be used when performing operations in a pipeline.

    ```
    cat mylist.txt | c8y inventory get --progress
    ```

### PSc8y improvements

* The `Data` parameter now accepts a path object (System.IO.FileSystemInfo). Previously the user would have to use the .Name or .FullName of the object.

    **Example**

    ```powershell
    $file = New-TemporaryFile
    '{"name": "myName in a file"}' | Out-File $file
    New-ManagedObject -Data $file
    ```

* `Wait-Operation` Fixed warning message format when operation fails. Id was being show in the incorrect position

## Released

### v1.11.0

#### New Features

* Exit codes are not set to the HTTP exit codes for commands which send a REST request. The https status codes are mapped to exit codes between 0 - 128 to ensure compatibility to different operating systems and applications.
    
    The full error codes can be found on the new [error handling](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/error-handling/) concept page. However a few include:

    * 0 => 0 (REST request was ok)
    * 1 => 401 (Unauthorized)
    * 3 => 403 (Forbidden)
    * 4 => 404 (Not found)
    * 9 => 409 (Conflict/duplicate)
    * 22 => 422 (Unprocessable entity / invalid request)
    * 50 => 500 (Internal server error)

* Added `Remove-ApplicationBinary` command

    **Example: Remove all application binaries related to an application**

    **PowerShell**

    ```powershell
    Remove-ApplicationBinary -Application 12345 -BinaryId 9876

    # Or remove all application binaries (except the active one) for an application
    Get-ApplicationBinaryCollection -Id 12345 | Remove-ApplicationBinary -Application 12345
    ```

    **Bash/zsh**

    ```sh
    c8y applications deleteApplicationBinary --application 12345 --binaryId 9876
    ```

#### New Features (PSc8y)

* Added support for saving meta information about the requests to the in-built PowerShell InformationVariable common parameter

    [Documentation - Request metrics](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/powershell-request-metrics/):

    **Example: Save request to a variable without sending the request**

    ```powershell
    New-Device -Name my-test -WhatIf -InformationVariable requestInfo -InformationAction SilentlyContinue
    $requestInfo
    ```

    *Output*

    ```powershell
    What If: Sending [POST] request to [https://example123.my-c8y.com/inventory/managedObjects]

    Headers:
    Accept: application/json
    Authorization: Basic asdfasfd........
    Content-Type: application/json
    User-Agent: go-client
    X-Application: go-client

    Body:
    {
    "c8y_IsDevice": {},
    "name": "my-test"
    }
    ```

    **Example: Get the response time of a request**

    ```powershell
    New-Device -Name my-test -InformationVariable requestInfo -InformationAction SilentlyContinue
    Write-Host ("Response took {0}" -f $requestInfo.MessageData.responseTime)
    ```

    *Output*

    ```
    Response took 172ms
    ```

* Support for ErrorVariable common variable to save error output to a variable

    **Example: Save error output to a variable**

    ```powershell
    Get-ManagedObject -Id 0 -ErrorVariable "c8yError" -ErrorAction SilentlyContinue

    if ($LASTEXITCODE -ne 0) {
        $MainError = $c8yError[-1]
        Write-Error "Something went wrong. details=$MainError"
    }
    ```


* Importing PSc8y will enforce Unicode (UTF-8) encoding on the console if the console is not already using UTF-8 (mostly affecting Windows as it does not use UTF-8 by default unlike MacOS and Linux)
    * User will be informed how to add the setting to the PowerShell profile
    * Enforcement of UTF-8 encoding can be disabled by setting

        ```powershell
        $env:C8Y_DISABLE_ENFORCE_ENCODING = $true
        ```
#### Minor Changes

* Updated PowerShell version from 7.0 to 7.1.1 inside docker image `c8y-pwsh`. This fixed a bug when using `Foreach-Object -Parallel` which would re-import modules instead of re-using it within each runspace.
* PSc8y will enforce PowerShell encoding to UTF8 to prevent encoding issues when sending data to the c8y go binary. The console encoding will be changed when importing `PSc8y`. UTF8 is the only text encoding supported. This mainly effects Windows, as MacOS and Linux use UTF8 encoding on the console by default.
* Added a global `--noColor` to the c8y binary to remove console colours from the log messages to make it easier to parse entries. By default PowerShell uses this option when calling the c8y binary as PowerShell handling the coloured log output itself.

#### Bug fixes

* `New-Device` fixed bug which prevent the command from creating the managed object
    * `Name` is no longer mandatory and it does not accepted piped input
* Changed the processing of standard output and error from the c8y binary to prevent read deadlocks when being called from PowerShell module PSc8y. #39


### v1.10.0

#### Breaking Changes

* `Expand-Device` no longer fetches the device managed object when given the id when being called from a function that does not makes use of a "Force" parameter. If you would like the old functionality, then add the new "-Fetch" parameter when calling `Expand-Device`.

    By default if Expand-Device is called from an interactive function, then the managed object will be looked up in order to provide helpful information to the user in the confirmation prompt. Additionally users writing modules, can force fetching of the device when given an id by using the new `-Fetch` parameter.

    The change enable a significant reduction in API calls as shown below in the following examples:

    **Comparison of total API calls per command to previous PSc8y version**

    Previous version: PSc8y=1.9.1

    The following examples show how many api calls are made to Cumulocity.

    ```sh
    # API calls: 1 x GET    (previously 3 x GET!)
    Get-Device 1234

    # API calls: 2 x GET    (previously 5 x GET!)
    Get-Device 1234 | Get-Device

    # API calls: 1 x GET and 1 x PUT    (prevously 4 x GET and 1 x PUT)
    Get-Device 1234 | Update-Device

    # API calls: 1 x POST   (prevously 3 x GET and 1 x POST)
    Add-DeviceToGroup -Group 11111 -NewChildDevice 222222 -ProcessingMode QUIESCENT -Force
    ```

    **Expand-Device Usage**
    Expand-Device was created in order to normalize the input of devices given by the user. Since PSc8y accepts devices either by id, name, object or piped objects, it can make it difficult to handle each of the input types in each function.

    ```powershell
    # file: my-script.ps1
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true
        )]
        [object[]] $Device = ""
    )

    foreach ($iDevice in (PSc8y\Expand-Device $Device)) {
        Write-Host ("Dummy api call with device id: /inventory/managedObject/{0}" -f $iDevice.id)
    }
    ```

    By writing it like this the user can call the function in the following ways with the same code.

    ```powershell
    # Pass array item
    ./my-script.ps1 -Device 12345

    # array of items mixing ids with names
    ./my-script.ps1 -Device "myDevicename", 1234

    # using pipelines from other PSc8y cmdlets
    Get-DeviceCollection | ./my-script.ps1

    # By hashtable with id property (using positional argument)
    ./my-script.ps1 @{id="12345"}, @{id="6789"}
    ```

    Now let's say that you wanted to add some logic which required the full device managed object from the server, and not just the id and name fields. This can be achieved by adding the `-Fetch` parameter to the `Expand-Device` cmdlet call.

    The differences in the output can be shown in the small example

    ```powershell
    # This will not fetch the device (as the device already exists)
    PS> 1234 | Expand-Device

    id   name
    ---- ----
    1234 [id=1234]
    ```

    Agent but using `-Fetch`.

    ```powershell
    # Using -Fetch will return the whole device managed object from Cumulocity (1 x GET request)
    PS> 1234 | Expand-Device -Fetch

    additionParents : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/additionParents}
    assetParents    : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/assetParents}
    c8y_IsDevice    : 
    childAdditions  : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/childAdditions}
    childAssets     : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/childAssets}
    childDevices    : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/childDevices}
    creationTime    : 1/19/2021 7:49:33 PM
    deviceParents   : @{references=System.Object[]; self=https://example.cumulocity.com/inventory/managedObjects/1234/deviceParents}
    id              : 1234
    lastUpdated     : 1/19/2021 8:52:29 PM
    name            : mynewname
    owner           : user@example.com
    self            : https://example.cumulocity.com/inventory/managedObjects/3882
    ```

#### New Features

* Added commands to manage managed object child additions

    **PowerShell**

    * `Get-ChildAdditionCollection`
    * `New-ChildAddition`
    * `Remove-ChildAddition`

    **Bash/zsh**

    * `c8y inventoryReferences listChildAdditions`
    * `c8y inventoryReferences createChildAddition`
    * `c8y inventoryReferences deleteChildAddition`

#### Performance improvements

* Reduced number of API calls within PSc8y and c8y binary by skipping lookups when an ID is given by the user. Previously PSc8y and c8y were sending two API calls to the server in order to normalize the request by retrieving additional information and potentiall shown to the user. Since this is currently not used, it has been removed.

#### Bug fixes

* `Set-Session` no longer causes the terminal bell/chime when using backspace or arrow keys.

### v1.9.1

#### Bug fixes

* `Get-C8ySessionProperty` selects first matching Session parameter in the call stack if multiple matches are found

### v1.9.0

### New features

* Added bulk operations commands

    **PowerShell**

    * `Get-BulkOperationCollection`
    * `Get-BulkOperation`
    * `New-BulkOperation`
    * `Update-BulkOperation`
    * `Remove-BulkOperation`
    
    **Bash/zsh**

    * `c8y bulkOperations list`
    * `c8y bulkOperations create`
    * `c8y bulkOperations get`
    * `c8y bulkOperations update`
    * `c8y bulkOperations delete`

* `Get-OperationCollection` supports `bulkOperationId` parameter to return operations related to a specific bulk operation id

* Added helpers function to make creating custom functions easier and which behave like native `PSc8y` cmdlets.
    * `Get-ClientCommonParameters` - Get common PSc8y parameters so they can be added to external using PowerShell's `DynamicParam` block
    * `Add-ClientResponseType` - Add a type to a list of devices if the `-Raw` parameter is not being used

    **Example**

    The following function get a list of software items stored as managed objects in Cumulocity.

    The cmdlet only needs to define one parameter $Name for the custom logic. The following parameters are inherited via the `Get-ClientCommonParameters` call in the DynamicParam block
    * Pagination parameters: PageSize, WithTotalPages, TotalPages, CurrentPage, IncludeAll
    * Pagination parameters: Session
    * General parameters: TimeoutSec, Raw, OutputFile

    ```powershell
    Function Get-SoftwareCollection {
        [cmdletbinding(
            SupportsShouldProcess = $true,
            ConfirmImpact = "None"
        )]
        Param(
            # Software name
            [string]
            $Name = "*"
        )
        # Inherit common parameters from PSc8y module
        DynamicParam { PSc8y\Get-ClientCommonParameters -Type "Collection" }

        Process {
            $Query = "type eq 'c8y_Software' and name eq '{0}'" -f $Name
            $null = $PSBoundParameters.Remove("Name")

            Find-ManagedObjectCollection -Query $Query @PSBoundParameters `
                | Select-Object `
                | Add-ClientResponseType -Type "application/vnd.com.nsn.cumulocity.customSoftware+json"
        }
    }
    ```

### Minor improvements

* "owner" is field is left untouched in the -Data parameter allowing the user to change it if required.
    ```powershell
    Update-ManagedObject -Id 12345 -Data @{owner="myuser"}
    ```

* Cumulocity API error messages are prefixed with "Server error." to make it more clear that the error is due to an API call and not the client.

    ```powershell
    PS /workspaces/go-c8y-cli> Get-AuditRecord 12345                        

    WARNING: c8y returned a non-zero exit code. code=1
    Write-Error: /workspaces/go-c8y-cli/tools/PSc8y/dist/PSc8y/PSc8y.psm1:3657:13
    Line |
    3657 |              Invoke-ClientCommand `
        |              ~~~~~~~~~~~~~~~~~~~~~~
        | Server error. general/internalError

    PS /workspaces/go-c8y-cli> Get-Alarm 12345                              


    WARNING: c8y returned a non-zero exit code. code=1
    Write-Error: /workspaces/go-c8y-cli/tools/PSc8y/dist/PSc8y/PSc8y.psm1:2742:13
    Line |
    2742 |              Invoke-ClientCommand `
        |              ~~~~~~~~~~~~~~~~~~~~~~
        | Server error. alarm/Not Found: Finding alarm from database failed : No alarm for gid '12345'!
    ```

* PSc8y command automatically detect the `-WhatIf` and `-Force` parameters from any parent commands. This reduces the amount of boilerplate code.

    **Example: Custom command to send a restart operation**

    ```powershell
    Function New-MyCustomRestartOperation {
        [cmdletbinding(
            SupportsShouldProcess = $true,
            ConfirmImpact = "High"
        )]
        Param(
            [Parameter(
                Mandatory = $true,
                Position = 0
            )]
            [object[]] $Device,

            [switch] $Force
        )

        Process {
            foreach ($iDevice in (Expand-Device $Device)) {
                New-Operation `
                    -Device $iDevice `
                    -Description "Restart device" `
                    -Data @{ c8y_Restart = @{}}
            }
        }
    }
    ```

    Normally when using `New-Operation` within your command, you need to pass the `WhatIf` and `Force` parameter values like so:
    
    ```powershell
    New-Operation `
        -Device $iDevice `
        -Data @{ c8y_Restart = @{}} `
        -WhatIf:$WhatIfPreference `
        -Force:$Force
    ```

    However now all PSc8y commands will automatically inherit these values.

    ```powershell
    New-MyCustomOperation -Device 12345 -WhatIf
    New-MyCustomOperation -Device 12345 -Force
    ```

    The variable inheritance can be disabled by setting the following environment variable

    ```powershell
    $env:C8Y_DISABLE_INHERITANCE = $true
    ```
* pwsh docker image improvements
    * Enabled pwsh tab completion by default
    * Added `vim` text editor

* Get-DeviceCollection supports `OrderBy` parameter to sort
    
    **Example: Get a list of devices sorting by name in descending order**

    **PowerShell**

    ```powershell
    Get-DeviceCollection -OrderBy "name desc"
    ```

    **Bash/zsh**
    
    ```sh
    c8y devices list --orderBy "name desc"
    ```

* Test cmdlets supports the `-Time` parameter to be able to control the timestamp of created entity. By default it will use "0s" (i.e. now). 
    * `New-TestAlarm`
    * `New-TestEvent`
    * `New-TestMeasurement`

* `Get-SessionHomePath` Added public PowerShell cmdlet to retrieve the path where the session are stored

* New cmdlet `Register-ClientArgumentCompleter` to enable other modules to add argument completion to PSc8y parameters like `Session` and `Template`
    * Note: `-Force` needs to be used if your command uses Dynamic Parameters

### v1.8.0

* `Get-Session` uses a new c8y session get to retrieve information about the current session
* Fixed bug when using the `-Session` on PUT and POST commands which resulted in an error being displayed eventhough the request would be successful
* `Expand-Device` supports piping of alarms, events, measurements and operations
* Added `-ProcessingMode` parameter to all commands that use DELETE, PUT and POST requests.

    ```powershell
    New-ManagedObject -Name myobject -ProcessingMode TRANSIENT
    New-ManagedObject -Name myobject -ProcessingMode QUIESCENT
    New-ManagedObject -Name myobject -ProcessingMode PERSISTENT
    New-ManagedObject -Name myobject -ProcessingMode CEP
    ```
* `Set-session` automatically selects a session if only one matching session is found rather than prompting the user for the selection
* `source` fragment is removed when being passed via file to the `Data` parameter in all create and update commands
    ```json
    // myevent.json
    {
        "source": {
            "id": "99999",
            "self": "https:/..../event/events/99999"
        },
        "type": "myExample1",

    }
    ```

    When executing the following command:
    ```powershell
    PSc8y\New-Event -Device 12345 -Data myevent.json
    ```

    The `source` id fragment will be replaced entirely by the new source as specified by the `Device` parameter

    ```json
    // myevent.json
    {
        "source": {
            "id": "12345",
        },
        "type": "myExample1",

    }
    ```
* Added logout user command to invalidate current user token

    **Bash/zsh**

    ```sh
    c8y users logout
    ```

    **PowerShell**

    ```powershell
    Invoke-UserLogout
    ```
* Fixed binary upload bugs with both `New-EventBinary` and `Update-EventBinary` which resulted in multipart form data being included in the binary information

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
