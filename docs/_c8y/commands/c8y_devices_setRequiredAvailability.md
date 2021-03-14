---
layout: manual
# permalink: /:path/:basename
category: devices
title: setRequiredAvailability
---
## c8y devices setRequiredAvailability

Set required availability

### Synopsis

Set the required availability of a device. Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.

```
c8y devices setRequiredAvailability [flags]
```

### Examples

```
$ c8y inventory setRequiredAvailability --device 12345 --interval 10
Set the required availability of a device by name to 10 minutes
        
```

### Options

```
      --device strings          Device ID (required) (accepts pipeline)
  -h, --help                    help for setRequiredAvailability
      --interval int            Interval in minutes (required)
      --processingMode string   Processing mode
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

