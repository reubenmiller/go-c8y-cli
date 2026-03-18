---
category: sessions
title: c8y sessions create
---
Create session

### Synopsis

Create a new Cumulocity session

```
c8y sessions create [flags]
```

### Examples

```
$ c8y sessions create --mode dev --host "https://mytenant.eu-latest.cumulocity.com"
Example 1: Create a DEV new session. Prompt for username and password

$ c8y sessions create \
    --mode qual \
	--host "https://mytenant.eu-latest.cumulocity.com"
	--username "myUser@me.com"
Create a new QA (QUAL) session prompting for password

$ c8y sessions create --mode prod --host "https://mytenant.eu-latest.cumulocity.com" --noStorage
Create a new production session where only only GET commands are enabled (with no password storage)

$ c8y sessions create --mode prod --host "https://localhost:443" --insecure
Create a session which points to a local api endpoint (most like an Cumulocity Edge instance)

$ c8y sessions create --mode dev --host example.cumulocity.com --loginType BROWSER
Create a session which uses SSO / OAUTH2 using Authorization Flow (via a local web browser)
Note: Requires the "Redirect to the user interface application" to be enabled in Cumulocity

c8y sessions create --mode dev --host example.cumulocity.com --loginType BROWSER --browserCallback localhost:8008/mycallback
Create a session which uses SSO / OAUTH2 using Authorization Flow (via a local web browser)
and define an explicit redirect/callback URI which is whitelisted in the SSO providers configuration
Note: Requires the "Redirect to the user interface application" to be enabled in Cumulocity

$ c8y sessions create --mode dev --host example.cumulocity.com --loginType DEVICE
Create a session with SSO / OAUTH2 Device Flow (RFC 8628)
		
```

### Options

```
      --allowInsecure            Allow insecure connection (e.g. when using self-signed certificates)
      --browserCallback string   Custom redirect URI for the browser authorization code flow, e.g. http://127.0.0.1:8080/callback
      --description string       Description about the session
      --encrypt                  Encrypt passwords and tokens (occurs when logging in)
  -h, --help                     help for create
      --host string              Host. .e.g. test.cumulocity.com. (required)
      --loginType string         Login Type, e.g. BASIC, OAUTH2_INTERNAL, OAUTH2, BROWSER, DEVICE, CERTIFICATE, NONE
      --mode string              Session mode which controls which commands are enabled by default
      --name string              Name of the session
      --noStorage                Don't store any passwords or tokens in the session file
      --noTenantPrefix           Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false
      --password string          Password. If left blank then you will be prompted for the password
      --prompt                   Force prompting of missing information
      --tenant string            Tenant ID
      --token string             Token
      --username string          Username (without tenant). (required)
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
      --retries int                Max number of attempts when a failed http call is encountered (default 3)
      --select stringArray         Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'
      --session string             Session configuration
      --sessionMode string         Override default session mode for a single command which would normally be disabled
  -P, --sessionPassword string     Override session password
  -U, --sessionUsername string     Override session username. i.e. peter or t1234/peter (with tenant)
      --silentExit                 Silent status codes do not affect the exit code
      --silentStatusCodes string   Status codes which will not print out an error message
      --timeout string             Request timeout duration, i.e. 60s, 2m (default "600s")
      --totalPages int             Total number of pages to get
  -v, --verbose                    Verbose logging
      --view string                Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError                  Errors will be printed on stdout instead of stderr
      --withTotalElements          Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages             Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                Number of workers (default 1)
```

