---
category: devices
title: c8y devices list
---
Get device collection

### Synopsis

Get a collection of devices based on filter parameters

```
c8y devices list [flags]
```

### Examples

```
$ c8y devices list --name "sensor*" --type myType
Get a collection of devices of type "myType", and their names start with "sensor"

$ c8y devices list --query "name eq '*sensor*' and creationTime.date gt '2021-04-02T00:00:00'"
Get devices which names containing 'sensor' and were created after 2021-04-02

$ echo -e "c8y_MacOS\nc8y_Linux" | c8y devices list --queryTemplate "type eq '%s'"
Get devices with type 'c8y_MacOS' then devices with type 'c8y_Linux' (using pipeline)

```

### Options

```
      --agents                       Only include agents.
      --availability string          Filter by c8y_Availability.status
      --fragmentType string          Filter by fragment type
      --group string                 Filter by group inclusion
  -h, --help                         help for list
      --lastMessageDateFrom string   Filter c8y_Availability.lastMessage from a specific date
      --lastMessageDateTo string     Filter c8y_Availability.lastMessage to a specific date
      --name string                  Filter by name
      --orderBy string               Order by. e.g. _id asc or name asc or creationTime.date desc (default "name")
      --owner string                 Filter by owner
      --query string                 Additional query filter (accepts pipeline)
      --queryTemplate string         String template to be used when applying the given query. Use %s to reference the query/pipeline input
      --type string                  Filter by type
      --withParents                  Include a flat list of all parents and grandparents of the given object
```

### Options inherited from parent commands

```
      --abortOnErrors int          Abort batch when reaching specified number of errors (default 10)
      --allowEmptyPipe             Don't fail when piped input is empty (stdin)
      --cache                      Enable cached responses
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

