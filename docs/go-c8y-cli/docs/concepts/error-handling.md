---
layout: default
category: Concepts
title: Error handling
---

## Exit codes

The c8y binary has the following exit codes to indicate the success of the command. The exit codes include either information about command usage errors or the HTTP status code returned by the Cumulocity API call associated with the command.

### General exit codes

The following exit codes are general codes which relate to the usage of c8y commands.

|Exit code|Type|Description|
|---|-|--|
|0|Success|No error occurred|
|2|Cancelled|User cancelled the command or did not confirm a prompt|
|99|Unexpected error|Unknown error|
|100|System Error|Unexpected error when processing a command. i.e. can not write to a file, parsing error etc.|
|101|User Error|User/command error such as invalid arguments|
|102|No session loaded|The user has not selected a Cumulocity session yet|
|103|BatchAbortedWithErrors|The batched job was aborted due to too many errors|
|104|BatchCompletedWithErrors|The batched job completed but has 1 or more errors|
|105|BatchJobLimitExceeded|The batched job was stopped due to exceeding the total of allowed jobs though no other errors occurred|
|106|Command timed out|The command took longer than the timeout setting|
|107|Invalid alias|The command took longer than the timeout setting|
|108|Decryption error|Error occurred when decrypting a session, i.e. invalid passphrase|

### HTTP status code errors

The HTTP status code have been mapped to exit codes. To ensure compatibility with multiple operating systems, the exit codes are limited to values between 0 - 128. Therefore the HTTP status codes have been mapped to the following exit codes.

Note: Only HTTP status codes between 400 and 599 are mapped to exit codes. HTTP status 200 and 201 are classed as ok, and therefore will return an exit code of 0.

|Exit code|Status code|Description|description|
|---|--|--|--|
|1|401|StatusUnauthorized401|Authentication has failed, or credentials were required but not provided.|
|3|403|Forbidden|You are not authorized to access the API.|
|4|404|Not Found|Resource not found at given location.|
|5|405|Method Not Allowed|The employed HTTP method cannot be used on this resource (e.g., using "POST" on a read-only resource).|
|9|409|Update Conflict or Duplicate|The entity already exists in the data source. or the entity already exists in the data source.|
|13|413|Execution Timeout|Query had been running too long and was timed out.|
|22|422|Invalid Data|General error with entity data format.|
|29|429|Too Many Requests|If the request rate limit per second is exceeded, the requests are delayed and kept in queue until the queue number limit is exceeded in which case the request is terminated with an error.|
|40|400|Bad Request|The request could not be understood by the server due to malformed syntax. The client SHOULD NOT repeat the request without modifications.|
|50|500|Internal Server Error|An internal error in the software system has occurred and the request could not be processed.|
|52|502|Bad Gateway|The server, while acting as a gateway or proxy, received an invalid response from the upstream server|
|53|503|Service Unavailable|The service is currently not available. This may be caused by an overloaded instance or it is down for maintenance. Please try it again in a few minutes.|

## Handling errors (Shell)

Errors which are related to command usage (not server errors), are written to standard error, while server responses (error or otherwise) are written to standard output.

For example, the following code tries to retrieve an non-existent managed object (id=0).

```bash
response=$( c8y inventory get --id=0 )
code=$?

if [ $code -ne 0 ]; then
  echo "An error occurred: code=$code"
else
  echo "ok"
fi
```

*Output*

```bash
2021-04-13T08:24:58.484+0200    ERROR   serverError: Finding device data from database failed : No managedObject for id '0'! GET https://test-tenant.example.com/inventory/managedObjects/0: 404 inventory/Not Found Finding device data from database failed : No managedObject for id '0'!
An error occurred: code=4
```


### Silencing specific HTTP status codes

Errors and warnings will be written to stderr. This means in the previous example the `response` variable will be empty, and an error message will still be displayed on the console (as no stderr redirection is being used).

You can silence specific error output for specific status codes by using the `silentStatusCodes=xxx,yyyy` option. The previous example can also be written as a one-liner:

```bash
c8y inventory get --id=0 --silentStatusCodes=404,409 && echo "ok" || echo "an error occurred code=$?"
```

```bash title="Output"
an error occurred code=4
```

The most common use-case for this is when you are creating something and you don't want to see an error message if the item already exists.

For example, let's say that are creating users but some users might already exist. You just want to be sure that after the script has been run the user is there (regardless when it was created).

```bash
user=$( c8y users create --userName=myexampleuser --template "{ password: _.Password() }" --silentStatusCodes 409 --select userName -o csv || c8y users get --id myexampleuser --select userName -o csv )
if [[ -n "$user" ]]; then
  echo "OK: $user: "
else
  echo "Unexpected error"
fi
```

### Include error in response

If you would like to return the error as json, then the `--withError` option can be given which will return any errors as a json response where individual fields can be parsed by using jq, or any other json parser.

```bash
# Note: stderr is redirected to null so it is not printed to the console
c8y inventory get --id=0 --withError 2>/dev/null | jq
```

```bash title="Output"
{
  "errorType": "serverError",
  "message": "Finding device data from database failed : No managedObject for id '0'!",
  "statusCode": 404,
  "exitCode": 4,
  "url": "/inventory/managedObjects/0",
  "c8yResponse": {
    "error": "inventory/Not Found",
    "message": "Finding device data from database failed : No managedObject for id '0'!",
    "info": "https://cumulocity.com/guides/reference/rest-implementation/#a-name-error-reporting-a-error-reporting"
  }
}
```

## Handling errors (PowerShell)

All PSc8y commands can be tested if they were successful or not by checking the PowerShell variable `$LASTEXITCODE`.

This variable stores the exit code from the last call to the binary (in this case the c8y binary).

So to check if a command was successful, the just check the `$LASTEXITCODE` value.

```powershell
New-ManagedObject -Name "my object" -Force

if ($LASTEXITCODE -ne 0) {
    Write-Error "Something went wrong!"
}
```

### Example: Handling specific errors based on the exit code

Let's say that your are writing a script to create an external identity for an existing device and you don't care if the external Id already exists (as it indicates that your script has already been run on this device).

Instead of first checking if the device has the external id, you can combine the requests into a single request and check the exit code for to determine if it already exists or was created.

The following shows an example of this using a file called `create-identity.ps1`.

```powershell
[cmdletbinding()]
Param(
  # Serial number. This must be unique!
  [Parameter(Mandatory = $true)]
  [string] $Device,

  # Serial number. This must be unique!
  [Parameter(Mandatory = $true)]
  [string] $SerialNumber
)

# try to create the new id (store the error messages, and silence them because we want to handle within our script)
$response = New-ExternalId `
  -Device $Device `
  -Type "mySerial" `
  -Name $SerialNumber `
  -WithError

# Save the code for better re-use (in case something else updates it unexpectedly)
$Code = $LASTEXITCODE

# Handle the different outcomes
switch ($Code) {
  0 {
    # 0 -> No errors
    Write-Host "Success: Created new identity. name=$SerialNumber" -ForegroundColor Green
  }

  9 {
    # 9 -> Status code 409 : Duplicate/conflict
    # In this case, this is still ok, but the message will be shown in a different colour
    Write-Host "Success: Identity already existed. name=$SerialNumber" -ForegroundColor Yellow
  }

  3 {
    # 3 -> Status code 403 : Permission denied
    Write-Error "Failed: You do not have the correct ROLE to create an external identity"
  }

  Default {
    # Handle unknown error
    $ErrorDetails = $c8yError[-1]
    Write-Error "Failed due to some other error! code=$_, details=$($response.message)"
  }
}
```

The script can be called using the following command

```powershell
./create-identity.ps1 -Device 12345 -SerialNumber 7DHD875d501SS
```

```powershell title="Output"
Success: Created new identity. name=7DHD875d501SS

Success: Identity already existed. name=7DHD875d501SS

Failed: You don not have the correct ROLE to create an external identity

Failed to Some other error occurred! code=22, details=.....
```
