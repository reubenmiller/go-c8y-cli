---
layout: default
title: Workers
---

import CodeExample from '@site/src/components/CodeExample';

By default go-c8y-cli uses 1 worker to process the api requests (a.k.a. jobs). The number of workers can be increased to speed up the processing of the jobs.

As with any concurrency, you should not just blindly increase the number of workers as this could have negative effects on your Cumulocity instance. You should start out conservative with any increases, and also use the given protection mechanisms such as delay, abort on error count, max jobs etc, to achieve better control over the process.

:::tip
The goal here is **controlled** concurrency, where requests are sent in a sustainable way and do not impact other activities / normal platform usage.
So use `workers <int>` with a suitable delay `delay <duration>` to limit any negative effects when processing large amounts of jobs.
:::

## Protection mechanisms

When sending large amounts of jobs (>1000), there are a few things to watch out for in order to be a good *api-citizen*.

* Total workers (number of workers which will process the jobs)
* Delay (time each worker should wait before retrieving the next job)
* Abort on error count (stop processing when the total errors reaches a specific amount)

### workers

The `workers` parameter controls the number of concurrent workers used to process the jobs. The more workers there are the quicker the requests will be processed, however the more load it puts on the platform.

:::info
You can protect against using too many workers via the configuration `settings.defaults.maxWorkers`

```sh
c8y settings update defaults.maxWorkers 10
```
:::

### delay

The delay parameter can be used to control how fast a worker retrieves the next job. It accepts a duration as a string, i.e.

* `--delay 10ms`
* `--delay 250ms`
* `--delay 1s`

### abortOnErrors

To prevent sending a large number of errors, go-c8y-cli will stop processing job once the total number of errors exceeds a specific value.

The `abortOnErrors <int>` parameter is used to control the threshold.

<CodeExample transform="false">

```bash
c8y util repeat 10 --format "%s%s" |
    c8y inventory get --abortOnErrors 1
```

</CodeExample>

```bash title="Output"
2021-05-24T11:33:40.046Z        ERROR   serverError: Finding device data from database failed : No managedObject for id '1'! GET https://{host}/inventory/managedObjects/1: 404 inventory/Not Found Finding device data from database failed : No managedObject for id '1'!
2021-05-24T11:33:40.049Z        ERROR   commandError: aborted batch as error count has been exceeded. totalErrors=1
```

## Progress indicator

The progress indicator can be used to track the status of long running commands.

The progress indicator shows the following information:

* start time
* elapsed time
* total requests sent
* current average worker request rate (realtime)


```json title="Output"
elapsed 00:06   (started: 2021-05-24T08:30:01Z)     ⠦ total requests sent:  140
worker 1:      ⠦                                           avg: 4.469 request/s
worker 2:      ⠦                                           avg: 4.449 request/s
worker 3:      ⠦                                           avg: 4.414 request/s
worker 4:      ⠦                                           avg: 4.348 request/s
worker 5:      ⠦                                           avg: 4.409 request/s
```

:::note
The average request rate of the workers only shows the request rate of recent commands so that it can give a rough indication of the current responsiveness of the platform.
:::

## Examples

### Adding a fragment to all devices

Image you have 50K devices and you wish to add a fragment to each of them. This means you need to send 50K requests (jobs), 1 request per device.

Normally this task is not so easy to do, and would require writing a custom script or application where you would have to worry about paging through devices, updating the fragment and adding concurrency to speed up the work.

However with go-c8y-cli, this is a simple one liner, by chaining two commands together:

1. Get all devices
2. Update each device with a fragment

The above steps look like this on the command line:

<CodeExample>

```bash
c8y devices list --includeAll |
    c8y devices update --template "{myFragment:{}}" --workers 5 --delay 200ms
```

</CodeExample>

The above command will use 5 workers when adding the fragment to each device, and wait 200ms before processing the next device (on a per worker basis). The `includeAll` parameter is being used so that go-c8y-cli will paginate through the entire device list, and not just the first N devices.

Since we are dealing with 50K devices, we should not assume that all of the requests will be processed without any errors. So ideally we should make the command idempotent so that we can run the command again, and it will only get the devices which have not yet been processed. Fortunately this is done by just adding a custom inventory query statement when retrieving the devices.

In addition, we'll add a timestamp inside your fragment to indicate when it was updated, so that we have a record in the platform when the request was sent.

<CodeExample>

```bash
c8y devices list --query "not(has(myFragment))" --includeAll |
    c8y devices update --template "{myFragment:{lastUpdated: _.Now()}}" --workers 5 --delay 200ms
```

</CodeExample>


### Setting device required availability

Changing the required interval across all devices is again a simple one liner. Though this time the progress indicator is being used to track the commands.

<CodeExample>

```bash
c8y devices list --includeAll |
    c8y devices setRequiredAvailability --interval 30 \
        --noAccept --processingMode QUIESCENT \
        --workers 5 --delay 200ms --progress
```

</CodeExample>

```json title="Output"
elapsed 00:06   (started: 2021-05-24T08:30:01Z)     ⠦ total requests sent:  140
worker 1:      ⠦                                           avg: 4.469 request/s
worker 2:      ⠦                                           avg: 4.449 request/s
worker 3:      ⠦                                           avg: 4.414 request/s
worker 4:      ⠦                                           avg: 4.348 request/s
worker 5:      ⠦                                           avg: 4.409 request/s
```

:::tip
`--noAccept` is used to indicate that we do not want the server to send us the updated device back, as we do not need it, so it can increase performance a bit because Cumulocity has less to do.
:::

:::tip
Using `--processingMode QUIESCENT` can be helpful to reduce load on the realtime processing engine by indicating that the update request should not go through it.
:::
