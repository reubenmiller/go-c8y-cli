---
category: Concepts - Extensions - API based commands
title: Body building
---

import CodeExample from '@site/src/components/CodeExample';

One major aspect of an API is the creation or modification of entities. Such entities could be any thing such as devices, applications, configuration etc.

API based commands provide an easy way for users to create the json bodies required for `PUT` and `POST` HTTP requests. The bodies can constructed by combining any of the following:


* Custom Flags
* Static templates (defined in the spec)
* Global `--template` flag
* Global `--data` flag


### Best practices

Some general pointers to keep in mind when generating the body via the commands line:

* Allow users to use `data` and `template` to create bodies (this gives them maximum flexibility)
* Keep number of flags below 10 (people generally like going through too many options)
* Provide templates if a request requires many fields or a very complex/nested data structure


## JSON Bodies

By default the body of `PUT` and `POST` requests will use a json format.

Below shows an api command which is used to provide a convenient way for users to create a new device using some custom fragments.

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


````bash title="Output (dry)"
What If: Sending [POST] request to [https://{host}/inventory/managedObjects]

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

## Binary

### Uploading a file (with meta information)

This kind of upload is called a `multipart/form-data` request. The request is generally used for uploading both binary files in addition to extra meta information describing the binary being uploaded.

It is typically used in Cumulocity IoT to add new inventory binaries or application binaries to the platform.

The file upload scenario can be utilized by using the special `file` type in the body section. The snippet below adds a `utils` group command with a single command called `upload`.

```yaml title="file: api/utils.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/extensionCommands.json
---
group:
  name: utils
  description: Example commands

commands:
  - name: upload
    description: Upload a file to the inventory man
    exampleList:
      - command: c8y examples utils upload --file ./myfile.txt
        description: Upload a single file
    method: POST
    path: inventory/binaries
    body:
      - name: file
        type: file          # <== the 'file' type activates the multipart form-data upload
        description: File
      
      - name: name
        type: string
        description: Set the name of the binary file. This will be the name of the file when it is downloaded in the UI

      - name: type
        type: string
        required: false
        description: Custom type. If left blank, the MIME type will be detected from the file extension
```

Assuming the api command was added to an extension called `examples`, the command can be executed to upload a file using:

<CodeExample transform="false">

```bash
c8y examples utils upload --file ./myfile.txt
```

</CodeExample>


Additional meta information can also be added by using the `template` or `data` flags which are automatically added to all requests which support a body.

<CodeExample transform="false">

```bash
c8y examples utils upload --file ./myfile.txt --template "{foo:{bar:'other data'}}"
```

</CodeExample>

Like any of the commands, the output of the command can be inspected by using the `dry` flag.


````markdown title="Output (dry)"
What If: Sending [POST] request to [https://{host}/inventory/binaries]

### POST /inventory/binaries

| header            | value
|-------------------|---------------------------
| Accept            | application/json 
| Authorization     | Bearer {token} 
| Content-Type      | multipart/form-data; boundary=9c365f1a9d792bcad0b782b551bd2ef0e1d3d7382a1d5de600a2c77999f9 

#### Body

```text
--9c365f1a9d792bcad0b782b551bd2ef0e1d3d7382a1d5de600a2c77999f9
Content-Disposition: form-data; name="file"; filename="myfile.txt"
Content-Type: application/octet-stream

testme

--9c365f1a9d792bcad0b782b551bd2ef0e1d3d7382a1d5de600a2c77999f9
Content-Disposition: form-data; name="object"

{
   "foo": {
      "bar": "other data"
   },
   "name": "myfile.txt",
   "type": "text/plain; charset=utf-8"
}

--9c365f1a9d792bcad0b782b551bd2ef0e1d3d7382a1d5de600a2c77999f9--

```
````

### Uploading binary data

Plain binary data can be sent in a request using the `fileContents` type. Using this type within the `body` section will send the raw file contents in the body. In contrast to the `file` type, it will not send the request as a multipart form-data request.

To demonstrate this, a new command can be added called `replace` which sends a PUT request to the inventory binary API to replace an existing binary with the contents from a new file. The api command definition to achieve this is shown below:

```yaml title="file: api/utils.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/extensionCommands.json
---
group:
  name: utils
  description: Example commands

commands:
  # ... previous command (omitted for simplicity)

  - name: replace
    description: Replace the contents of a binary file with new contents
    method: PUT
    path: inventory/binaries/{id}
    pathParameters:
      - name: id
        type: string
        description: Existing binary id
    body:
      - name: file
        type: fileContents      # <== Just upload bytes
        description: file
```

This command will require us to know the id of the existing binary that we wish to replace, however it should be easy enough to find using the other `go-c8y-cli` command (e.g. `c8y binaries list`).

<CodeExample transform="false">

```bash
c8y examples utils replace --id 12345 --file ./myfile.txt --dry
```

</CodeExample>

Using `dry` shows how the request body is different to the previous example which used a multipart form-data upload.

````bash title="Output (dry)"
What If: Sending [PUT] request to [https://{host}/inventory/binaries/12345]

### POST /inventory/binaries

| header            | value
|-------------------|---------------------------
| Accept            | application/json 
| Authorization     | Bearer {token} 
| Content-Type      | application/json 

#### Body

```text
testme

```
````
