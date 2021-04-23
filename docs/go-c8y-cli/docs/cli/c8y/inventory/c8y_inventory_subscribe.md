---
category: inventory
title: c8y inventory subscribe
---
Subscribe to realtime managedObjects

### Synopsis

Subscribe to realtime managedObjects

```
c8y inventory subscribe [flags]
```

### Examples

```
$ c8y inventory subscribe --device 12345
Subscribe to managedObjects (in realtime) for device 12345

$ c8y inventory subscribe --device 12345 --duration 30
Subscribe to managedObjects (in realtime) for device 12345 for 30 seconds

$ c8y inventory subscribe --count 10
Subscribe to managedObjects (in realtime) for all devices, and stop after receiving 10 managedObjects
		
```

### Options

```
      --actionTypes strings   Realtime action types, i.e. CREATE,UPDATE,DELETE
      --count int             Max number of realtime notifications to wait for
      --device string         Device ID
      --duration int          Timeout in seconds (default 30)
  -h, --help                  help for subscribe
```

### Options inherited from parent commands

```
      --abortOnErrors int          Abort batch when reaching specified number of errors (default 10)
  -c, --compact                    Compact instead of pretty-printed output. Pretty print is the default if output is the terminal
      --confirm                    Prompt for confirmation
      --confirmText string         Custom confirmation text
      --currentPage int            Current page size which should be returned
      --debug                      Set very verbose log messages
      --delay int                  delay in milliseconds after each request
      --delayBefore int            delay in milliseconds before each request
      --dry                        Dry run. Don't send any data to the server
      --dryFormat string           Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --filter strings             filter
      --flatten                    flatten
  -f, --force                      Do not prompt for confirmation
  -H, --header strings             custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
      --includeAll                 Include all results by iterating through each page
  -l, --logMessage string          Add custom message to the activity log
      --maxJobs int                Maximum number of jobs. 0 = unlimited (use with caution!)
      --noAccept                   Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
  -M, --noColor                    Don't use colors when displaying log entries on the console
      --noLog                      Disables the activity log for the current command
      --noProxy                    Ignore the proxy settings
  -o, --output string              Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string          Save JSON output to file (after select)
      --outputFileRaw string       Save raw response to file
  -p, --pageSize int               Maximum results per page (default 5)
      --progress                   Show progress bar. This will also disable any other verbose output
      --proxy string               Proxy setting, i.e. http://10.0.0.1:8080
      --queryParam strings         custom query parameters. i.e. --queryParam "withCustomOption=true,myOtherOption=myvalue"
  -r, --raw                        Show raw response. This mode will force output=json and view=off
      --select stringArray         select
      --session string             Session configuration
  -P, --sessionPassword string     Override session password
  -U, --sessionUsername string     Override session username. i.e. peter or t1234/peter (with tenant)
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout float              Timeout in seconds (default 600)
      --totalPages int             Total number of pages to get
  -v, --verbose                    Verbose logging
      --view string                View option (default "auto")
      --withError                  Errors will be printed on stdout instead of stderr
  -t, --withTotalPages             Include all results
      --workers int                Number of workers (default 1)
```

