---
layout: manual
# permalink: /:path/:basename
category: sessions
title: create
---
## c8y sessions create

Create session

### Synopsis

Create a new Cumulocity session

```
c8y sessions create [flags]
```

### Examples

```
### Example 1: Create a DEV new session. Prompt for username and password

$ c8y sessions create --type dev --host "https://mytenant.eu-latest.cumulocity.com"

### Example 2: Create a new QA (QUAL) session prompting for password

$ c8y sessions create \
    --type qual \
	--host "https://mytenant.eu-latest.cumulocity.com"
	--username "myUser@me.com"

### Example 3: Create a new production session where only only GET commands are enabled (with no password storage)

$ c8y sessions create --type prod --host "https://mytenant.eu-latest.cumulocity.com" --noStorage
		
```

### Options

```
      --description string   Description about the session
      --encrypt              Encrypt passwords and tokens (occurs when logging in)
  -h, --help                 help for create
      --host string          Host. .e.g. test.cumulocity.com. (required)
      --name string          Name of the session
      --noStorage            Don't store any passwords or tokens in the session file
      --noTenantPrefix       Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false
      --password string      Password. If left blank then you will be prompted for the password
      --tenant string        Tenant ID
      --token string         Token
      --type string          Session type. List of predefined session types
      --username string      Username (without tenant). (required)
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

