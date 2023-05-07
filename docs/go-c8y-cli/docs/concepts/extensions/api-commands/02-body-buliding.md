---
category: Tutorials - Extensions
title: Body building
---

A HTTP request's body can be built by using the following components:

* Flags
* Static templates
* Global `--template` flag
* Global `--data` flag


### Body

```yaml
commands:
  - name: create
    description: Create device
    method: POST
    path: inventory/managedObjects
    body:
      - name: name
        type: string
        description: Device name
        required: true
        pipeline: true

      - name: type
        type: string
        property: type
        description: Device type
        validationSet:
          - linux
          - windows
          - macos
      
      - name: interval
        type: integer
        property: c8y_RequiredAvailability.responseInterval
        description: Response interval (for availability monitoring)

      - name: option1
        type: boolean
        description: Add static value if the flag is used. If not used, then nothing will be added
        value: customValue1

      - name: agent
        type: optional_fragment
        property: com_cumulocity_model_Agent
        description: Add special c8y agent fragment

      
    bodyTemplates:
      - type: jsonnet
        template: |
          {c8y_IsDevice:{}}
```

<CodeExample>

```sh
c8y organizer assets create --name foo --type macos --dry --interval 10 --agent --option1
```

</CodeExample>


````bash title="Output"
What If: Sending [POST] request to [https://test-ci-runner01.latest.stage.c8y.io/inventory/managedObjects]

### POST /inventory/managedObjects

| header            | value
|-------------------|---------------------------
| Accept            | application/json 
| Authorization     | Bearer {token} 
| Content-Type      | application/json 

#### Body

```json
{
  "c8y_IsDevice": {},
  "c8y_RequiredAvailability": {
    "responseInterval": 10
  },
  "com_cumulocity_model_Agent": {},
  "name": "foo",
  "option1": "customValue1",
  "type": "macos"
}
```
````


### Non-json bodies

TODO

### Uploading binaries

TODO
