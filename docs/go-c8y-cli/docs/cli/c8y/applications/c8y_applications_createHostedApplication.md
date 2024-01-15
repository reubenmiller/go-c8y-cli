---
category: applications
title: c8y applications createHostedApplication
---
Create hosted application

### Synopsis

Create a new hosted web application or update the binary of an existing hosted application

If the zip file or folder contains a 'cumulocity.json' manifest file, then the key will be automatically read from it.


```
c8y applications createHostedApplication [flags]
```

### Examples

```
$ c8y applications createHostedApplication --file ./myapp.zip
Create new hosted application from a given zip file. The application will be called "myapp". If the application placeholder does not exist then it will be created

$ c8y applications createHostedApplication --file simple-helloworld/build --name custom_helloworld
Create/update hosted web application from a build folder and specify a custom application name

$ c8y applications createHostedApplication --file myapp.zip --skipActivation
Create/update hosted web application but don't activate it, so the current version (if any) will be untouched

```

### Options

```
      --availability string        Access level for other tenants. Possible values are : MARKET, PRIVATE (default)
      --contextPath string         contextPath of the hosted application
  -d, --data stringArray           static data to be applied to body. accepts json or shorthand json, i.e. --data 'value1=1,my.nested.value=100'
      --file string                File or Folder of the web application. It should contain a index.html file in the root folder/ or zip file
  -h, --help                       help for createHostedApplication
      --key string                 Shared secret of application. Defaults to the value inside the cumulocity.json file (if present)
      --name string                Name of application
      --resourcesUrl string        URL to application base directory hosted on an external server. Required when application type is HOSTED (default "/")
      --skipActivation             Don't activate to the application after it has been created and uploaded
      --skipUpload                 Don't uploaded the web app binary. Only the application placeholder will be created
      --template string            Body template
      --templateVars stringArray   Body template variables
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
  -P, --sessionPassword string     Override session password
  -U, --sessionUsername string     Override session username. i.e. peter or t1234/peter (with tenant)
      --silentExit                 Silent status codes do not affect the exit code
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout string             Request timeout duration, i.e. 60s, 2m (default "60s")
      --totalPages int             Total number of pages to get
  -v, --verbose                    Verbose logging
      --view string                Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError                  Errors will be printed on stdout instead of stderr
      --withTotalElements          Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```

