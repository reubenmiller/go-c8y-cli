---
category: assert
title: c8y devices assert fragments
---
Assert fragments on a device

### Synopsis

Assert fragments on a device by polling until a condition is met or a timeout is reached

```
c8y devices assert fragments [flags]
```

### Examples

```
$ c8y devices assert fragments --id 1234 --fragment "c8y_Mobile.iccd"
# Fragments for the managed object to have a non-null c8y_Mobile.iccd fragment

$ c8y devices assert fragments --id 1234 --fragment '!c8y_Mobile'
# Fragments for the managed object to not have a c8y_Mobile fragment

$ c8y devices assert fragments --id 1234 --fragment 'name=^\d+-\w+$'
# Fragments for the managed object name fragment to match the regular expression '^\d+-\w+'

$ c8y devices assert fragments --id 1234 --fragment 'name=^$' --fragment c8y_IsDevice
# Fragments for the managed object name fragment to match the regular expression '^\d+-\w+'

$ c8y devices assert fragments --id 1234 --duration 1m --interval 10s
# Fragments for the operation to be set to SUCCESSFUL and give up after 1 minute

$ c8y devices list --device 1111 | c8y operations fragments --status "FAILED" --status "SUCCESSFUL"
# Fragments for operation to be set to either FAILED or SUCCESSFUL

```

### Options

```
      --device strings      The ManagedObject which is the source of this event. (accepts pipeline)
      --duration string     Timeout duration. i.e. 30s or 1m (1 minute) (default "30s")
      --fragments strings   Fragments to fragments for. If multiple values are given, then it will be applied as an OR operation
  -h, --help                help for fragments
      --interval string     Interval to check on the status, i.e. 10s or 1min (default "5s")
      --retries int         Number of retries before giving up per id
      --strict              Strict mode, fail if no match is found
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
      --noProxy                    Ignore the proxy settings
  -n, --nullInput                  Don't read the input (stdin). Useful if using in shell for/while loops
  -o, --output string              Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string          Save JSON output to file (after select/view)
      --outputFileRaw string       Save raw response to file (before select/view)
  -p, --pageSize int               Maximum results per page (default 5)
      --progress                   Show progress bar. This will also disable any other verbose output
      --proxy string               Proxy setting, i.e. http://10.0.0.1:8080
  -r, --raw                        Show raw response. This mode will force output=json and view=off
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
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```

