---
layout: manual
# permalink: /:path/:basename
category: c8y
title: rest
---
## c8y rest

Send custom request

### Synopsis

Send a custom REST request to a given endpoint

```
c8y rest [flags]
```

### Examples

```
### Get a list of alarms
$ c8y rest GET /alarm/alarms

### Get a list of alarms with custom query parameters
c8y rest GET "/alarm/alarms?pageSize=10&status=ACTIVE"

### Create a new alarm
c8y rest POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={'id': '12345'}"
		
```

### Options

```
      --accept string         accept (header)
      --contentType string    content type (header)
  -d, --data string           json
      --file string           File to be uploaded as a binary
  -H, --header strings        headers. i.e. --header "Accept: value"
  -h, --help                  help for rest
      --host string           host to use for the rest request. If empty, then the session's host will be used
      --template string       Body template
      --templateVars string   Body template variables
```

### Options inherited from parent commands

```
      --abortOnErrors int          Abort batch when reaching specified number of errors (default 10)
  -c, --compact                    Compact instead of pretty-printed output. Pretty print is the default if output is the terminal
      --compress                   Alias for --compact for users coming from PowerShell
      --confirmText string         Custom confirmation text
      --currentPage int            Current page size which should be returned
      --debug                      Set very verbose log messages
      --delay int                  delay in milliseconds after each request (default 1000)
      --dry                        Dry run. Don't send any data to the server
      --dryFormat string           Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --filter strings             filter
      --flatten                    flatten
  -f, --force                      Do not prompt for confirmation
      --includeAll                 Include all results by iterating through each page
  -l, --logMessage string          Add custom message to the activity log
      --maxJobs int                Maximum number of jobs. 0 = unlimited (use with caution!) (default 100)
      --noAccept                   Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
  -M, --noColor                    Don't use colors when displaying log entries on the console
      --noLog                      Disables the activity log for the current command
      --noProxy                    Ignore the proxy settings
  -o, --output string              Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string          Save JSON output to file (after select)
      --outputFileRaw string       Save raw response to file
  -p, --pageSize int               Maximum results per page (default 5)
      --progress                   Show progress bar. This will also disable any other verbose output
      --prompt                     Prompt for confirmation
      --proxy string               Proxy setting, i.e. http://10.0.0.1:8080
  -r, --raw                        Raw values
      --select stringArray         select
      --session string             Session configuration
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout float              Timeout in seconds (default 600)
      --totalPages int             Total number of pages to get
      --useEnv                     Allow loading Cumulocity session setting from environment variables
  -v, --verbose                    Verbose logging
      --view string                View file
      --withError                  Errors will be printed on stdout instead of stderr
  -t, --withTotalPages             Include all results
      --workers int                Number of workers (default 1)
```

