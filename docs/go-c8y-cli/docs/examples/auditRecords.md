---
title: Audit Records
---

import CodeExample from '@site/src/components/CodeExample';

## Get

### Retrieve the audit records related to an alarm

The history of an alarm can be recalled by using the following: 

<CodeExample>

```bash
c8y alarms get --id 70657 | c8y auditrecords list
```

</CodeExample>


```plaintext title="Output"
| id          | time                          | type       | AuditSourceDevice | activity           | text                                                      | user          | source.id   |
|-------------|-------------------------------|------------|-------------------|--------------------|-----------------------------------------------------------|---------------|-------------|
| 497752      | 2021-04-25T12:33:27.741Z      | Alarm      | 497835            | Alarm created      | Device name: 'device01', alarm text: 'Unknown error'      | ciuser01      | 497751      |
```

### Retrieve the audit records related to an operation

The history of an operation can be recalled by using the following:

<CodeExample>

```bash
c8y operations get --id 497931 | c8y auditrecords list
```

</CodeExample>

```plaintext title="Output"
| id          | time                          | type           | AuditSourceDevice | activity               | text                                                                                  | user          | source.id   |
|-------------|-------------------------------|----------------|-------------------|------------------------|---------------------------------------------------------------------------------------|---------------|-------------|
| 497932      | 2021-04-25T12:47:30.920Z      | Operation      | 497835            | Operation created      | Operation created: status='PENDING', description='Restart device', device name='devi… | ciuser01      | 497931      |
| 497933      | 2021-04-25T12:48:40.149Z      | Operation      | 497835            | Operation updated      | Operation updated: status='EXECUTING', description='Restart device', device name='de… | ciuser01      | 497931      |
| 497934      | 2021-04-25T12:48:50.694Z      | Operation      | 497835            | Operation updated      | Operation updated: status='SUCCESSFUL', description='Restart device', device name='d… | ciuser01      | 497931      |
```


## Create

### Create audit record related to a device

<CodeExample>

```bash
c8y auditrecords create \
    --type "ManagedObject" \
    --time "0s" \
    --text "Managed Object updated: my_Prop: value" \
    --source "497835" \
    --activity "Managed Object updated" \
    --severity "information"
```

</CodeExample>


```plaintext title="Output"
| id          | time                          | type               | com_cumulocity_model_event_auditsourcedevice.id | activity                    | text                                        | user          | source.id   |
|-------------|-------------------------------|--------------------|-------------------------------------------------|-----------------------------|---------------------------------------------|---------------|-------------|
| 497753      | 2021-04-25T12:52:17.580Z      | ManagedObject      |                                                 | Managed Object updated      | Managed Object updated: my_Prop: value      | ciuser01      | 497835      |
```
