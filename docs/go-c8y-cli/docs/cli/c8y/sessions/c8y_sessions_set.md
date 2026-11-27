---
category: sessions
title: c8y sessions set
---
Set Cumulocity session

### Synopsis

Set a session, login and test the session and get either OAuth2 token, or using two factor authentication

```
c8y sessions set [flags]
```

### Examples

```
$ eval $( c8y sessions set )
Set a session interactively

$ eval $( c8y sessions set --sessionFilter "company dev" )
Set a session but only include session matching company AND dev

$ eval $( c8y sessions set --session myfile.json --tfaCode 123456 )
Set a session using a given file (non-interactively)

$ eval $( c8y sessions set --session myfile.json --loginType DEVICE )
Login using the OAuth2 device flow (polls for approval after visiting the URL shown in the terminal)

$ eval $( c8y sessions set --session myfile.json --loginType BROWSER )
Login using Authorization Code flow – opens the system browser and waits for the callback

$ eval $( c8y sessions set --session myfile.json --loginType BROWSER --browserCallback http://127.0.0.1:8080/callback )
Login via browser with a custom callback URL registered in the SSO provider

$ eval $( c8y sessions set --session myfile.json --loginType CERTIFICATE )
Login using a device certificate (mTLS); reads cert paths from the session file or C8Y_CERTIFICATE / C8Y_CERTIFICATE_KEY env vars

```

### Options

```
      --browserCallback string   Custom redirect URI for the browser flow local callback server, e.g. http://127.0.0.1:8080/callback. Must match a URI registered in the SSO provider. Defaults to http://127.0.0.1:5001/callback
      --clear[=true|false]       Clear any existing tokens
      --loginType string         Login type preference, e.g. OAUTH2_INTERNAL, DEVICE (device flow), BROWSER (Authorization Code via browser), BASIC or CERTIFICATE (mTLS). OAUTH2 is a legacy alias for DEVICE. When set to BASIC, any existing token will be cleared
      --no-banner[=true|false]   Don't show the session banner
      --sessionFilter string     Filter to be applied to the list of sessions even before the values can be selected
      --shell string             Shell type to return the environment variables
      --tfaCode string           Two Factor Authentication code
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
      --paginationStrategy string        Pagination strategy used when iterating results (e.g. with --includeAll): auto (default), offset, id, time
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

