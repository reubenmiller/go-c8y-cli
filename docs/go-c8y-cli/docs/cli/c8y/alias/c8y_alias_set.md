---
category: alias <command>
title: c8y alias set
---
Create a shortcut for a c8y command

### Synopsis

Declare a word as a command alias that will expand to the specified command(s).
The expansion may specify additional arguments and flags. If the expansion
includes positional placeholders such as '$1', '$2', etc., any extra arguments
that follow the invocation of an alias will be inserted appropriately.
If '--shell' is specified, the alias will be run through a shell interpreter (sh). This allows you
to compose commands with "|" or redirect with ">". Note that extra arguments following the alias
will not be automatically passed to the expanded expression. To have a shell alias receive
arguments, you must explicitly accept them using "$1", "$2", etc., or "$@" to accept all of them.
Platform note: on Windows, shell aliases are executed via "sh" as installed by Git For Windows. If
you have installed git on Windows in some other way, shell aliases may not work for you.
Quotes must always be used when defining a command as in the examples.


```
c8y alias set <alias> <expansion> [flags]
```

### Examples

```
$ c8y alias set createTestDevice 'devices create --template test.device.json'
$ c8y createTestDevice
#=> c8y devices create --template test.device.json

$ c8y alias set listAlarmsByType 'alarms list --type="$1"'
$ c8y listAlarmsByType myType
#=> c8y alarms list --type="myType"

$ c8y alias set recentEvents 'operations list --dateFrom="-1h"'
$ c8y recentEvents
#=> c8y operations list --dateFrom="-1h"

$ c8y alias set --shell findInAudit 'c8y auditrecords list --dateFrom="-30d" --view auto --includeAll | grep $1'
$ c8y findInAudit deleted
#=> c8y auditrecords list --dateFrom="-30d" --output json --view auto --includeAll | grep deleted

```

### Options

```
  -h, --help    help for set
  -s, --shell   Declare an alias to be passed through a shell interpreter
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

