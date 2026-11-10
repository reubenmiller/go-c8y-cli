---
category: mcp
title: c8y mcp
---
Model Context Protocol (MCP) server for AI agents

### Synopsis

Run an MCP server which gives AI agents controlled access to a Cumulocity tenant using go-c8y-cli commands

### Options

```
  -h, --help[=true|false]   help for mcp
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

