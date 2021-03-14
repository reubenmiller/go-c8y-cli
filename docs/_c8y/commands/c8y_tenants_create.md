---
layout: manual
# permalink: /:path/:basename
category: tenants
title: create
---
## c8y tenants create

Create tenant

### Synopsis

Create a new tenant

```
c8y tenants create [flags]
```

### Examples

```
$ c8y tenants create --company "mycompany" --domain "mycompany" --adminName "admin" --password "mys3curep9d8"
Create a new tenant (from the management tenant)
        
```

### Options

```
      --adminName string        Username of the tenant administrator
      --adminPass string        Password of the tenant administrator
      --company string          Company name. Maximum 256 characters (required)
      --contactName string      A contact name, for example an administrator, of the tenant
      --contactPhone string     An international contact phone number
  -d, --data string             json
      --domain string           Domain name to be used for the tenant. Maximum 256 characters (required) (accepts pipeline)
  -h, --help                    help for create
      --processingMode string   Processing mode
      --template string         Body template
      --templateVars string     Body template variables
      --tenantId string         The tenant ID. This should be left bank unless you know what you are doing. Will be auto-generated if not present.
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

