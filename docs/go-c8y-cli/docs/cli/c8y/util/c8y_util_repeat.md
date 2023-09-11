---
category: util
title: c8y util repeat
---
Repeat input

### Synopsis

Generic utility to repeat input values x times

```
c8y util repeat [flags]
```

### Examples

```
$ c8y util repeat 5 --input "my name"
Repeat input value "my name" 5 times

$ echo "my name" | c8y util repeat 2 --format "my prefix - %s"
Repeat input value "my name" 2 times (using pipeline)
	=> my prefix - my name
	=> my prefix - my name

$ echo "device" | c8y util repeat 2 --offset 100 --format "%s %05s"
Repeat input value "device" 2 times (using pipeline)
	=> device 00101
	=> device 00102

$ c8y util repeat --infinite | c8y api --url "/service/report-agent/health" --raw --delay 1s
Use repeat to create an infinite loop, to check the health of a microservice waiting 1 seconds after each request

$ echo "device" | c8y util repeat 2 | c8y util repeat 3 --format "%s_%s"
Combine two calls to iterator over 3 devices twice. This can then be used to input into other c8y commands
	=> device_1
	=> device_2
	=> device_3
	=> device_1
	=> device_2
	=> device_3

$ c8y devices get --id 1235 | c8y util repeat 5 | c8y events create --text "test event" --type "myType" --dry --delay 1000ms
Get a device, then repeat it 5 times in order to create 5 events for it (delaying 1000 ms between each event creation)

$ c8y devices get --id 1234 | c8y util repeat 5 --randomDelayMin 1000ms --randomDelayMax 10000ms -v | c8y events create --text "test event" --type "myType"
Create 10 events for the same device and use a random delay between 1000ms and 10000ms between the creation of each event

$ echo "test" | c8y util repeat 5 --randomDelayMax 10000ms -v
Print "test" 5 times waiting between 0s and 10s after each line

$ echo "test" | c8y util repeat 5 --randomDelayMin 5s -v
Print "test" 5 times waiting exactly 5 seconds after each line

$ echo "test" | c8y util repeat --min 1 --max 10
Print "test" a random number of times, between 1 to 10 times (inclusive)

```

### Options

```
      --first int               only include first x lines. 0 = all lines
      --format string           format string to be applied to each input line (default "%s")
  -h, --help                    help for repeat
      --infinite                Repeat infinitely. You will need to ctrl-c it to stop it
      --input string            input value to be repeated (required) (accepts pipeline)
      --max int                 max number of (randomized) times to repeat the input (inclusive). 0 = no output (default 1)
      --min int                 min number of (randomized) times to repeat the input (inclusive) (default 1)
      --offset int              offset the output index counter. default = 0.
      --randomDelayMax string   random maximum delay after each request, i.e. 5ms, 1.2s. It must be >= randomDelayMin. 0 = disabled. (default "0ms")
      --randomDelayMin string   random minimum delay after each request, i.e. 5ms, 1.2s. It must be less than randomDelayMax. 0 = disabled (default "0ms")
      --randomSkip float32      randomly skip line based on a percentage, probability as a float: 0 to 1, 1 = always skip, 0 = never skip, -1 = disabled (default -1)
      --skip int                skip first x input lines
      --times int               number of times to repeat the input (default 1)
      --useLineCount            Use line count for the index instead of repeat counter
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
      --noProgress                 Disable progress bars
      --noProxy                    Ignore the proxy settings
  -n, --nullInput                  Don't read the input (stdin). Useful if using in shell for/while loops
  -o, --output string              Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string          Save JSON output to file (after select/view)
      --outputFileRaw string       Save raw response to file (before select/view)
      --outputTemplate string      jsonnet template to apply to the output
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
      --withTotalElements          Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```

