---
layout: default
category: Concepts
title: Error handling
---

### Exit codes

The c8y binary has the following exit codes to indicate the success of the command. The exit codes include either information about command usage errors or the HTTP status code returned by the Cumulocity API call associated with the command.

#### General exit codes

The following exit codes are general codes which relate to the usage of c8y commands.

|Exit code|Type|Description|
|---|-|--|
|0|Success|No error occurred|
|100|System Error|Unexpected error when processing a command. i.e. can not write to a file, parsing error etc.|
|101|User Error|User/command error such as invalid arguments|
|102|No session loaded|The user has not selected a Cumulocity session yet|

#### Exit code to HTTP status code errors

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

### Handling errors (Bash/zsh)

Errors which are related to command usage (not server errors), are written to standard error, while server responses (error or otherwise) are written to standard output.

For example, the following code tries to retrieve an non-existant managed object (id=0).

```sh
response=$( c8y inventory get --id=0 )
code=$?

if [ $code -ne 0 ]; then
  echo "An error occurred: code=$code, details=$response"
else
  echo "ok"
fi
```

Since the id does not exist, the server response will stored in the `response` variable. Therefore it can be used to display an error to the user.

Here is the same example but which just prints a simple message of the command was succesful or not (without error details).

```sh
c8y inventory get --id=0 > /dev/null  && echo "api call ok" || echo "api call failed"
```

The output can also be redirected to file for later, for example:

```sh
c8y inventory get --id=0 > response  && echo "success" || echo "failed. details=$(cat response | jq -r .message)"
```

*Output*

```sh
failed. details=Finding device data from database failed : No managedObject for id '0'!
```

### Handling errors in PowerShell

All PSc8y commands can be tested if they were successful or not by checking the PowerShell variable `$LASTEXITCODE`.

This variable stores the exit code from the last call to the binary (in this case the c8y binary).

So to check if a command was successful, the just check the `$LASTEXITCODE` value.

```powershell
New-ManagedObject -Name "my object" -Force

if ($LASTEXITCODE -ne 0) {
    Write-Error "Something went wrong!"
}
```

#### Saving errors to a variable

When using the PowerShell common parameter `ErrorVariable`, all errors that occured during the execution of the command will be stored in a variable with the given name. You can check the variable for detailed information about any errors.

Note: The `ErrorVariable` parameter accepts just the name of the variable not the actual variable itself!

For example, we try to get a managed object and store the error information to variable called "c8yError"

```powershell
$managedObject = Get-ManagedObject -Id 0 -ErrorVariable "c8yError"
```

You can check if the call was successful by checking the `$LASTEXITCODE` variable.

```powershell
$EverythingOK = $LASTEXITCODE -eq 0
```

In most cases you could just check the return value of the command, i.e. `$managedObject` as it only receives a value if the command was successful.

For this same example, we could re-write the check as:

```powershell
$EverythingWorked = $null -ne $managedObject
```

Note: The `$null` is on the left side of the "not equal" operator (-ne), because of the way PowerShell handles comparison between two objects. Using $null on the left side will ensure that you are checking if the object is $null and not the items within the array (should the $managedObjects be an array)

Now let's say that you didn't want to just check if the command was successful, but you wanted to check what kind of error occured (i.e. server error or a command/client error). To achieve this, all you have to do is check the last item in the `c8yError` array like so:

```powershell
$managedObject = Get-ManagedObject -Id 0 -ErrorVariable "c8yError"

if ($null -ne $managedObject) {
  #
  # Recveived managed object
  #
  Write-Host ("Found managed object: id={0}, name={1}" -f $managedObject.id, $managedObject.name)

} else {
  #
  # Custom error handling
  #
  switch -Regex ($c8yError[-1]) {
    "serverError" {
      Write-Error "Deteched server error. details=$_"
    }
    Default {
      Write-Error "Detected command/client error. details=$_"
    }
  }
}
```

Now the example above will now display two errors to the user. The first one is from PSc8y itself, and the second one from our custom error handling.

If you want to hide the original error, you can add the in-built PowerShell variable `-ErrorAction SilentlyContinue` to the `Get-ManagedObject` call:

```powershell
$managedObject = Get-ManagedObject -Id 0 -ErrorVariable "c8yError" -ErrorAction "SilentlyContinue"
```

The `c8yError` also has additional information for a detailed anaylsis of what went wrong. It stores the full verbose output of the c8y binary, so it can be helpful to look through it for additional clues to what went wrong.

#### Example: Handling specific errors based on the exit code

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
$null = New-ExternalId `
  -Device $Device `
  -Type "mySerial" `
  -Name $SerialNumber `
  -ErrorAction "SilentlyContinue" `
  -ErrorVariable "c8yError"

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
    Write-Error "Failed: You don not have the correct ROLE to create an external identity"
  }

  Default {
    # Handle unknown error
    $ErrorDetails = $c8yError[-1]
    Write-Error "Failed to Some other error occurred! code=$_, details=$ErrorDetails"
  }
}
```

The script can be called using the following command

```powershell
./create-identity.ps1 -Device 12345 -SerialNumber 7DHD875d501SS
```

*Outputs*

```powershell
Success: Created new identity. name=7DHD875d501SS

Success: Identity already existed. name=7DHD875d501SS

Failed: You don not have the correct ROLE to create an external identity

Failed to Some other error occurred! code=22, details=.....
```
