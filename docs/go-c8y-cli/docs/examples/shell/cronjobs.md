---
title: Cronjobs
---

Cronjobs are commonly used linux to run a task/script periodically. This page provides a few examples and important tips to keep in mind when using go-c8y-cli within a cronjob.


## Tips

The cron daemon uses a different environment to your normal shell instance, so there are a few things to watch out for when creating a cronjob which uses `go-c8y-cli`.

* Ensure the `c8y` binary is within the `PATH` environment variable. By default cron only adds `/usr/bin` and `/usr/sbin` to `PATH`. Check where `c8y` is installed by running `which c8y`.
* Always redirect stdin to `/dev/null` in the cronjob. This prevents stdin detection errors in go-c8y-cli which can be hard to debug
* Consider using a lock file (or similar mechanism) within the script being called in the cronjob to prevent multiple instances running at the same time


## Examples

### Running a script periodically as a cronjob

Run a script which cancels executing operations which are older than 2 days

```sh title="file: /opt/c8y/my-example-cronjob.sh"
#!/bin/bash
# Set CI mode which disables all confirmations
export C8Y_SETTINGS_CI=true

c8y operations list --dateTo -2d --status EXECUTING --includeAll \
| c8y operations update --status FAILED --failureReason "Cancelled stale operation"
```

```sh title="crontab -e"
# Make sure the PATH includes. Use 'which c8y' to check where the c8y binary is installed
PATH=/usr/local/bin:/usr/bin:/usr/sbin
SHELL=/bin/bash

* * * * * C8Y_SESSION=/opt/c8y/my-session.json /opt/c8y/my-example-cronjob.sh < /dev/null
```

:::caution
Standard input (`stdin`) should be redirected to `/dev/null` so that the automatic pipe detection of go-c8y-cli can work. Otherwise go-c8y-cli may incorrectly think there is piped input, due to the way cron sets `stdin` when calling the cronjob.

If you don't redirect stdin to `/dev/null` then you will have to use `-n / --nullInput` to tell c8y that it should ignore stdin.
:::

## Other useful tools

### Flock (file lock)

Flock is a nice tool to manage lock files from within shell scripts or the command line. These lock files are useful for knowing whether or not a script is running.

#### Example 1: Using flock directly in the cronjob definition

```sh title="crontab -e"
PATH=/usr/local/bin:/usr/bin:/usr/sbin
SHELL=/bin/bash

* * * * * /usr/bin/flock --timeout=1 /path/to/cron.lock /opt/c8y/myscript.sh < /dev/null
```

#### Example 2: Using flock within the bash script

Instead of calling flock from within the cronjob definition, it is called from inside the bash script. This has the advantage because the locking will also be respected regardless on how it is being called (i.e. manually or by cron).

```sh title="crontab -e"
PATH=/usr/local/bin:/usr/bin:/usr/sbin
SHELL=/bin/bash

* * * * * /opt/c8y/myscript.sh < /dev/null
```

The script below shows the two 
```sh title="file: /opt/c8y/myscript.sh"
#!/bin/bash
 
exec {lock_fd}>/var/lock/mylockfile || exit 1
flock -n "$lock_fd" || { echo "ERROR: flock() failed." >&2; exit 1; }

# ... commands executed under lock ...
c8y devices list

# Optional: Manually release the lock...(if left out, the lock will be removed when the script exits)
flock -u "$lock_fd"
```
