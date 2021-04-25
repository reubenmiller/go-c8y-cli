---
title: Audit Records
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Get

### Retrieve the audit records related to an alarm

The history of an alarm can be recalled by using the following: 

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms get --id 70657 | c8y auditrecords list
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Alarm -Id 70657 | Get-AuditRecordCollection
```

</TabItem>
</Tabs>

```plaintext title="Output"
| id          | time                          | type       | AuditSourceDevice | activity           | text                                                      | user          | source.id   |
|-------------|-------------------------------|------------|-------------------|--------------------|-----------------------------------------------------------|---------------|-------------|
| 497752      | 2021-04-25T12:33:27.741Z      | Alarm      | 497835            | Alarm created      | Device name: 'device01', alarm text: 'Unknown error'      | ciuser01      | 497751      |
```

### Retrieve the audit records related to an operation

The history of an operation can be recalled by using the following:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y operations get --id 497931 | c8y auditrecords list
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Operation -Id 497931 | Get-AuditRecordCollection
```

</TabItem>
</Tabs>

```plaintext title="Output"
| id          | time                          | type           | AuditSourceDevice | activity               | text                                                                                  | user          | source.id   |
|-------------|-------------------------------|----------------|-------------------|------------------------|---------------------------------------------------------------------------------------|---------------|-------------|
| 497932      | 2021-04-25T12:47:30.920Z      | Operation      | 497835            | Operation created      | Operation created: status='PENDING', description='Restart device', device name='devi… | ciuser01      | 497931      |
| 497933      | 2021-04-25T12:48:40.149Z      | Operation      | 497835            | Operation updated      | Operation updated: status='EXECUTING', description='Restart device', device name='de… | ciuser01      | 497931      |
| 497934      | 2021-04-25T12:48:50.694Z      | Operation      | 497835            | Operation updated      | Operation updated: status='SUCCESSFUL', description='Restart device', device name='d… | ciuser01      | 497931      |
```


## Create

### Create audit record related to a device

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y auditrecords create \
    --type "ManagedObject" \
    --time "0s" \
    --text "Managed Object updated: my_Prop: value" \
    --source "497835" \
    --activity "Managed Object updated" \
    --severity "information"
```

</TabItem>
<TabItem value="powershell">

```powershell
New-AuditRecord `
    -Type "ManagedObject" `
    -Time "0s" `
    -Text "Managed Object updated: my_Prop: value" `
    -Source "497835" `
    -Activity "Managed Object updated" `
    -Severity "information"
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | time                          | type               | com_cumulocity_model_event_auditsourcedevice.id | activity                    | text                                        | user          | source.id   |
|-------------|-------------------------------|--------------------|-------------------------------------------------|-----------------------------|---------------------------------------------|---------------|-------------|
| 497753      | 2021-04-25T12:52:17.580Z      | ManagedObject      |                                                 | Managed Object updated      | Managed Object updated: my_Prop: value      | ciuser01      | 497835      |
```
