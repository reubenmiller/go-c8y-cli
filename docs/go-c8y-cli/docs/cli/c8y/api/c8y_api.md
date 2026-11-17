---
category: c8y
title: c8y api
---
Send api request

### Synopsis

Send an authenticated api request to a given endpoint

```
c8y api [flags]
```

### Examples

```
$ c8y api GET /alarm/alarms
Get a list of alarms

$ c8y api GET "/alarm/alarms&status=ACTIVE" --pageSize 10
Get a list of alarms with custom query parameters

$ c8y api GET "/dosomething" --host localhost:8080
Send a GET request to a custom hosted service running on localhost port 8080. You can force https or http by adding them as a prefix to the host.

$ c8y api GET "/dosomething" --host $C8Y_HOST/service/mymicroservice --dry
Send a request to a microservice hosted in the cloud. The /dosomething will be appended to the /service/mymicroservice path provided by the host flag.
Results in the following request: GET /service/mymicroservice/dosomething sent to the  current session URL.

$ c8y api GET "/alarm/alarms&status=ACTIVE" --pageSize 1 --withTotalPages
Get a total ACTIVE alarms

$ c8y api POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source.id='12345'" --keepProperties
Create a new alarm

$ c8y activitylog list --filter "method like POST" | c8y api --method DELETE
Get items created via POST from the activity log and delete them 

$ echo -e "/inventory/1111\n/inventory/2222" | c8y api --method PUT --template "{myScript: {lastUpdated: _.Now() }}"
Pipe a list of urls and execute HTTP PUT and use a template to generate the body

$ echo "12345" | c8y api PUT "/service/example" --template "{id: input.value}"
Send a PUT request to a fixed url, but use the piped input to build the request's body

$ echo "12345" | c8y api PUT "/service/example/%s" --template "{id: input.value}"
Send a PUT request using the piped input in both the url and the request's body ('%s' will be replaced with the current piped input line)

$ echo "{\"url\": \"/service/custom/endpoint\",\"body\":{\"name\": \"peter\"}}" | c8y api POST --template "input.value.body"
Build a custom request using piped json. The input url property will be mapped to the --url flag, and use
a template to also build the request's body from the piped input data.

$ c8y api POST /example -F file=@./mydevice01.pem -F parentAssetId=1111 -F assetTypeId=2222
Send a multipart form-data request with a file and custom form fields (similar to curl's -F option)

$ c8y api POST /example --form "file=@./data.bin;type=application/octet-stream;filename=custom.bin" --form "description=<./notes.txt"
Upload a file with a custom content type and filename, and set a form field to the contents of a file

```

### Options

```
      --accept string                 accept (header)
      --contentType string            content type (header)
  -d, --data stringArray              static data to be applied to body. accepts json or shorthand json, i.e. --data 'value1=1,my.nested.value=100'
      --file string                   File to be uploaded as a binary
  -F, --form stringArray              Multipart form field (curl style). Use name=value for text, name=@path[;type=mime/type][;filename=name] for a file, or name=<path for text read from a file. Can be used multiple times. Implies POST if no method is given
      --formdata string               form data (json or shorthand json)
      --host string                   host to use for the rest request. If empty, then the session's host will be used
      --keepProperties[=true|false]   Don't strip Cumulocity properties from the data property, i.e. source etc. (default true)
      --method string                 HTTP method (default "GET")
      --template string               Body template
      --templateVars stringArray      Body template variables
      --url string                    URL path. Any reference to '%s' will be replaced with the current input value (accepts pipeline)
```

### Options inherited from parent commands

```
      --abortOnErrors int                Abort batch when reaching specified number of errors (default 10)
      --allowEmptyPipe[=true|false]      Don't fail when piped input is empty (stdin)
      --cache[=true|false]               Enable cached responses
      --cacheBodyPaths strings           Cache should limit hashing of selected paths in the json body. Empty indicates all values
      --cacheTTL string                  Cache time-to-live (TTL) as a duration, i.e. 60s, 2m (default "60s")
  -c, --compact[=true|false]             Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal
      --confirm[=true|false]             Prompt for confirmation
      --confirmText string               Custom confirmation text
      --currentPage int                  Current page which should be returned
      --customQueryParam strings         add custom URL query parameters. i.e. --customQueryParam 'withCustomOption=true,myOtherOption=myvalue'
      --debug[=true|false]               Set very verbose log messages
      --delay string                     delay after each request, i.e. 5ms, 1.2s (default "0ms")
      --delayBefore string               delay before each request, i.e. 5ms, 1.2s (default "0ms")
      --dry[=true|false]                 Dry run. Don't send any data to the server
      --dryFormat string                 Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --examples[=true|false]            Show examples for the current command
      --filter stringArray               Apply a client side filter to response before returning it to the user
      --flatten[=true|false]             flatten json output by replacing nested json properties with properties where their names are represented by dot notation
  -f, --force[=true|false]               Do not prompt for confirmation. Ignored when using --confirm
  -H, --header strings                   custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
  -h, --help[=true|false]                Show help. Optionally give a level to limit the flags shown: all, local, global, category (list the categories), or one or more categories, i.e. --help local,output
      --includeAll[=true|false]          Include all results by iterating through each page
  -k, --insecure[=true|false]            Allow insecure server connections when using SSL
  -l, --logMessage string                Add custom message to the activity log
      --maxJobs int                      Maximum number of jobs. 0 = unlimited (use with caution!)
      --noAccept[=true|false]            Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
      --noCache[=true|false]             Force disabling of cached responses (overwrites cache setting)
  -M, --noColor[=true|false]             Don't use colors when displaying log entries on the console
      --noLog[=true|false]               Disables the activity log for the current command
      --noProgress[=true|false]          Disable progress bars
      --noProxy[=true|false]             Ignore the proxy settings
  -n, --nullInput[=true|false]           Don't read the input (stdin). Useful if using in shell for/while loops
  -o, --output string                    Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string                Save JSON output to file (after select/view)
      --outputFileRaw string             Save raw response to file (before select/view)
      --outputTemplate string            jsonnet template to apply to the output
  -p, --pageSize int                     Maximum results per page (default 5)
      --progress[=true|false]            Show progress bar. This will also disable any other verbose output
      --proxy string                     Proxy setting, i.e. http://10.0.0.1:8080
  -r, --raw[=true|false]                 Show raw response. This mode will force output=json and view=off
      --retries int                      Max number of attempts when a failed http call is encountered (default 3)
      --select stringArray               Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'
      --session string                   Session configuration
      --sessionMode string               Override default session mode for a single command which would normally be disabled
  -P, --sessionPassword string           Override session password
  -U, --sessionUsername string           Override session username. i.e. peter or t1234/peter (with tenant)
      --silentExit[=true|false]          Silent status codes do not affect the exit code
      --silentStatusCodes string         Status codes which will not print out an error message
      --timeout string                   Request timeout duration, i.e. 60s, 2m (default "600s")
      --totalPages int                   Total number of pages to get
  -v, --verbose[=true|false]             Verbose logging
      --view string                      Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError[=true|false]           Errors will be printed on stdout instead of stderr
      --withTotalElements[=true|false]   Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages[=true|false]      Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                      Number of workers (default 1)
```

