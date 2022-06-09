---
title: shell
---

## Intro

1. select a session

2. Check Cumulocity version

3. Show raw output or using jq

    ```bash
    c8y currenttenant version

    c8y currenttenant version | jq
    ```

## Create a single device

```bash
c8y devices create --name demo01
```

Display device with dry to check before creating it

```bash
c8y devices create --name demo02 --template device.complex.jsonnet --dry
```

## Create a lot of devices

```bash
c8y util repeat 100 --format "device_%s%04s" | c8y devices create --name demo02 --template device.complex.jsonnet --workers 5 --delay 100
```

## Check for errors in the activity log

```sh
c8y activitylog
```

Check in the audit log

```bash
c8y auditrecords list --dateFrom "-100min" --select id,source.id,activity
```

## Run a custom inventory query

```bash
c8y devices list --query "creationTime.date gt '2021-05-01T00:00:00'"
```


## Update a lot of devices

```bash
c8y devices list --includeAll | 
```


## Periodically create 10 alarms

```
nohup c8y devices get --id 497814 | c8y util repeat 10 --delayBefore 5000 | c8y alarms create --template test.alarm.jsonnet --force > /dev/null &

c8y alarms subscribe --device 497814
```

## Add

c8y alarms subscribe --device 501771 --duration 1000 --count 1 &

c8y alarms create --device demo01 --template alarm.jsonnet --type "myType" --text "Example alarm"

## Subscribe to alarms

c8y alarms subscribe --device 501771 --duration 1000 --count 1 | c8y alarms update --status ACKNOWLEDGED --delayBefore 2000 -f &

Subscribe

c8y alarms create --device demo01 --template alarm.jsonnet --type "myType" --text "Example alarm"

