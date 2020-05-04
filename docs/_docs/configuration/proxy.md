---
layout: default
category: Configuration
title: Proxy
---

c8y has first class support for corporate proxies. The proxy can either be configured globally by using environment variables, or by using arguments.

### Setting proxy using environment variables

##### Bash

```sh
export HTTP_PROXY=http://10.0.0.1:8080
export NO_PROXY=localhost,127.0.0.1,edge01.server
```

##### Powershell

```sh
$env:HTTP_PROXY = "http://10.0.0.1:8080"
$env:NO_PROXY = "localhost,127.0.0.1,edge01.server"
```

### Overriding proxy settings for individual commands

The proxy environment variables can be ignore for individual commands by using the `--noProxy` option.


```sh
c8y devices list --noProxy
```

Or, the proxy settings can be set for a single command by using the  `--proxy` argument:

```sh
c8y devices list --proxy "http://10.0.0.1:8080
```
