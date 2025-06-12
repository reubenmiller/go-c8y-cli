---
category: deviceregistration
title: c8y deviceregistration register-external-ca
---
Register device with an x509 certificate from an external certificate authority

### Synopsis

The certificate should already be issued to the device by the external CA and the CA certificate
should be added to Cumulocity as a Trusted Certificate.

This is required when the Trusted Certificate has the autoRegistration setting disabled, so it will
require a device to be registered with the platform before the certificate is accepted by Cumulocity


```
c8y deviceregistration register-external-ca [flags]
```

### Examples

```
$ c8y deviceregistration register-external-ca --id "ASDF098SD1J10912UD92JDLCNCU8"
Register a new device which which is signed by an external ca

$ echo -e "device1\ndevice2" | c8y deviceregistration register-external-ca --type linux --template "{name: input.value}"
Register 2 devices, and set the names based on their external id (using CERTIFICATES auth)
     
```

### Options

```
  -d, --data stringArray           static data to be applied to body. accepts json or shorthand json, i.e. --data 'value1=1,my.nested.value=100'
      --external-type string       The type of the external ID. If IDTYPE doesn't appear in the file, the default value is used. The default value is c8y_Serial (default "c8y_Serial")
      --group string               The path in the groups hierarchy where the device is added. PATH contains the name of each group separated by /, that is: main_group/sub_group/.../last_sub_group. If a group does not exist, the import creates the group
  -h, --help                       help for register-external-ca
      --iccid string               The ICCID of the device (SIM card number). If the ICCID appears in file, the import adds a fragment c8y_Mobile.iccid.
      --id strings                 Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)
      --name string                Device name. Defaults to the external id
      --processingMode string      Cumulocity processing mode
      --template string            Body template
      --templateVars stringArray   Body template variables
      --tenant string              The ID of the tenant for which the registration is executed (only allowed for the management tenant)
      --type string                Device type (default "thin-edge.io")
```

### Options inherited from parent commands

```
      --abortOnErrors int          Abort batch when reaching specified number of errors (default 10)
      --allowEmptyPipe             Don't fail when piped input is empty (stdin)
      --cache                      Enable cached responses
      --cacheBodyPaths strings     Cache should limit hashing of selected paths in the json body. Empty indicates all values
      --cacheTTL string            Cache time-to-live (TTL) as a duration, i.e. 60s, 2m (default "60s")
  -c, --compact                    Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal
      --confirm                    Prompt for confirmation
      --confirmText string         Custom confirmation text
      --currentPage int            Current page which should be returned
      --customQueryParam strings   add custom URL query parameters. i.e. --customQueryParam 'withCustomOption=true,myOtherOption=myvalue'
      --debug                      Set very verbose log messages
      --delay string               delay after each request, i.e. 5ms, 1.2s (default "0ms")
      --delayBefore string         delay before each request, i.e. 5ms, 1.2s (default "0ms")
      --dry                        Dry run. Don't send any data to the server
      --dryFormat string           Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --examples                   Show examples for the current command
      --filter stringArray         Apply a client side filter to response before returning it to the user
      --flatten                    flatten json output by replacing nested json properties with properties where their names are represented by dot notation
  -f, --force                      Do not prompt for confirmation. Ignored when using --confirm
  -H, --header strings             custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
      --includeAll                 Include all results by iterating through each page
  -k, --insecure                   Allow insecure server connections when using SSL
  -l, --logMessage string          Add custom message to the activity log
      --maxJobs int                Maximum number of jobs. 0 = unlimited (use with caution!)
      --noAccept                   Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
      --noCache                    Force disabling of cached responses (overwrites cache setting)
  -M, --noColor                    Don't use colors when displaying log entries on the console
      --noLog                      Disables the activity log for the current command
      --noProgress                 Disable progress bars
      --noProxy                    Ignore the proxy settings
  -n, --nullInput                  Don't read the input (stdin). Useful if using in shell for/while loops
  -o, --output string              Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string          Save JSON output to file (after select/view)
      --outputFileRaw string       Save raw response to file (before select/view)
      --outputTemplate string      jsonnet template to apply to the output
  -p, --pageSize int               Maximum results per page (default 5)
      --progress                   Show progress bar. This will also disable any other verbose output
      --proxy string               Proxy setting, i.e. http://10.0.0.1:8080
  -r, --raw                        Show raw response. This mode will force output=json and view=off
      --retries int                Max number of attempts when a failed http call is encountered (default 3)
      --select stringArray         Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'
      --session string             Session configuration
      --sessionMode string         Override default session mode to allow once-off commands which would normally be disabled
  -P, --sessionPassword string     Override session password
  -U, --sessionUsername string     Override session username. i.e. peter or t1234/peter (with tenant)
      --silentExit                 Silent status codes do not affect the exit code
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout string             Request timeout duration, i.e. 60s, 2m (default "600s")
      --totalPages int             Total number of pages to get
  -v, --verbose                    Verbose logging
      --view string                Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError                  Errors will be printed on stdout instead of stderr
      --withTotalElements          Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```

