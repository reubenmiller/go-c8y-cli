---
layout: manual
# permalink: /:path/:basename
category: alias <command>
title: set <alias> <expansion>
---
## c8y alias set

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
  -c, --compact                    Compact instead of pretty-printed output. Pretty print is the default if output is the terminal
      --confirm                    Prompt for confirmation
      --confirmText string         Custom confirmation text
      --currentPage int            Current page size which should be returned
      --debug                      Set very verbose log messages
      --delay int                  delay in milliseconds after each request (default 1000)
      --delayBefore int            delay in milliseconds before each request
      --dry                        Dry run. Don't send any data to the server
      --dryFormat string           Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --filter strings             filter
      --flatten                    flatten
  -f, --force                      Do not prompt for confirmation
  -H, --header strings             custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
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
      --proxy string               Proxy setting, i.e. http://10.0.0.1:8080
      --queryParam strings         custom query parameters. i.e. --queryParam "withCustomOption=true myOtherOption=myvalue"
  -r, --raw                        Raw values
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
