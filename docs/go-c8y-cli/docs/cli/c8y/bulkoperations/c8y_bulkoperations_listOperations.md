---
category: bulkoperations
title: c8y bulkoperations listOperations
---
Get operations collection

### Synopsis

Get a collection of operations related to a bulk operation

```
c8y bulkoperations listOperations [flags]
```

### Examples

```
$ c8y bulkoperations listOperations --id 10 --status PENDING
Get a list of pending operations from bulk operation with id 10

$ c8y bulkoperations list --filter "status eq IN_PROGRESS" | c8y bulkoperations listOperations --status PENDING
Get all pending operations from all bulk operations which are still in progress (using pipeline)

$ c8y bulkoperations list | c8y bulkoperations listOperations --status EXECUTING --dateTo "-10d" | c8y operations update --status FAILED --failureReason "Manually cancelled stale operation"
Check all bulk operations if they have any related operations still in executing state and were created more than 10 days ago, then cancel it with a custom message
        
```

### Options

```
      --dateFrom string   Start date or date and time of operation.
      --dateTo string     End date or date and time of operation.
  -h, --help              help for listOperations
      --id string         Bulk operation id. (required) (accepts pipeline)
      --revert            Sort operations newest to oldest. Must be used with dateFrom and/or dateTo parameters
      --status string     Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.
```

### Options inherited from parent commands

```
      --abortOnErrors int          Abort batch when reaching specified number of errors (default 10)
  -c, --compact                    Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal
      --confirm                    Prompt for confirmation
      --confirmText string         Custom confirmation text
      --currentPage int            Current page which should be returned
      --debug                      Set very verbose log messages
      --delay int                  delay in milliseconds after each request
      --delayBefore int            delay in milliseconds before each request
      --dry                        Dry run. Don't send any data to the server
      --dryFormat string           Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --examples                   Show examples for the current command
      --filter strings             Apply a client side filter to response before returning it to the user
      --flatten                    flatten json output by replacing nested json properties with properties where their names are represented by dot notation
  -f, --force                      Do not prompt for confirmation. Ignored when using --confirm
  -H, --header strings             custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
      --includeAll                 Include all results by iterating through each page
  -l, --logMessage string          Add custom message to the activity log
      --maxJobs int                Maximum number of jobs. 0 = unlimited (use with caution!)
      --noAccept                   Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
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
      --queryParam strings         custom query parameters. i.e. --queryParam "withCustomOption=true,myOtherOption=myvalue"
  -r, --raw                        Show raw response. This mode will force output=json and view=off
      --select stringArray         Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'
      --session string             Session configuration
  -P, --sessionPassword string     Override session password
  -U, --sessionUsername string     Override session username. i.e. peter or t1234/peter (with tenant)
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout float              Request timeout in seconds (default 600)
      --totalPages int             Total number of pages to get
  -v, --verbose                    Verbose logging
      --view string                Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError                  Errors will be printed on stdout instead of stderr
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statitics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```
