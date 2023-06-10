---
title: Importing data from csv
---

import CodeExample from '@site/src/components/CodeExample';

### Example: Import data from csv

**Scenario**

A customer provides some data exported by a 3rd party service that contains some data that should be converted to Cumulocity IoT events.

**Goal**

Create a Cumulocity IoT event for each for from a csv file.

**Procedure**

1. Let's assume we have a single csv file called `machine_events.csv`. We are going to convert each row to an individual event on the same device using a fixed type.

    Below shows the contents of the csv file that we will be using in this tutorial.

    ```csv title="machine_events.csv"
    timestamp,text
    1686391200,Alarm Reset drücken AKTIV<br>Alarm Rücktür AAA offen AKTIV
    1686391800,Alarm Reset drücken INAKTIV<br>Alarm Rücktür AAA offen INAKTIV
    1686392400,Alarm Reset drücken GELÖSCHT<br>Alarm Rücktür AAA offen GELÖSCHT
    1686393000,Alarm Sattelheizung1: Stoppen aktiv AKTIV
    ```

    The `machine_events.csv` contains just two columns; `timestamp` and `text`. However you may notice that the `timestamp` column contains a timestamp a unix timestamp (e.g. number of second since 1970-01-01), so this means that we will have to convert this to a Cumulocity IoT compatible timestamp (e.g. [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601)), but we will worry about this conversion later on.

2. We can check if the csv data is being read correctly and that it can be converted to json as the output will be the input mechanism used by downstream commands.

    CSV files can be read using the `c8y util repeatcsv` command. This commands is able to read a csv files and convert into a pipe-friendly format (json lines).

    Below shows shows an example of this and just outputs the result to the console for now to verify that the input data is processed correctly.

    <CodeExample>

    ```bash
    c8y util repeatcsv machine_events.csv
    ```

    </CodeExample>

    ```bash title="Output"
    | text                                                            | timestamp       |
    |-----------------------------------------------------------------|-----------------|
    | Alarm Reset drücken AKTIV<br>Alarm Rücktür AAA offen AKTIV      | 1686391200      |
    | Alarm Reset drücken INAKTIV<br>Alarm Rücktür AAA offen INAK…    | 1686391800      |
    | Alarm Reset drücken GELÖSCHT<br>Alarm Rücktür AAA offen GE…     | 1686392400      |
    | Alarm Sattelheizung1: Stoppen aktiv AKTIV                       | 1686393000      |
    ```

    :::note
    The output is shown as a table because the command's output is not being piped anywhere, so the it assumes that a human is viewing the output, and thus a table format is generally easier to read then json. You can change the output format by manually providing the `--output json` flag.
    :::

3. Now that we have confirmed that the csv data can be read correctly, we can pipe the data to the `c8y events create` command which will create one event per streamed input.

    However the input data needs to be shaped before the events can be created. We will use the `template` flag to shape the data into the required format. Since we are still defining the correct template, we will limit the amount of csv rows that are processed by using the `--first 1` flag until we get the template correct.

    <CodeExample>
    
    ```bash
    c8y util repeatcsv machine_events.csv --first 1 \
    | c8y events create \
        --template "{time: _.Now(input.value.timestamp), text: input.value.text, type: 'machine_CustomEvent'}" \
        --device 12345 \
        --dry
    ```

    </CodeExample>

    ````bash title="Output"
    What If: Sending [POST] request to [https://example.c8y.io/event/events]

    ### POST /event/events

    | header            | value
    |-------------------|---------------------------
    | Accept            | application/json 
    | Authorization     | Bearer {token} 
    | Content-Type      | application/json 

    #### Body

    ```json
    {
        "source": {
            "id": "12345"
        },
        "text": "Alarm Reset drücken AKTIV\u003cbr\u003eAlarm Rücktür AAA offen AKTIV",
        "time": "2023-06-10T12:00:00.000+02:00",
        "type": "machine_CustomEvent"
    }
    ```
    ````

    Though let's take a moment to understand what is happening in the `template`. The following shows the same template but it is reformatted to make it a bit more readable.

    ```js
    {
        time: _.Now(input.value.timestamp),
        text: input.value.text,
        type: 'machine_CustomEvent'
    }
    ```

    Below describes the logic behind each property:

    * The `time` property value is built using the `_.Now()` function to converts the unix timestamp from the piped input object (`input.value`) to an ISO 8601 timestamp.
    * The `text` just uses the `.text` data untouched from the input object (`input.value`)
    * The `type` uses a static string

4. Now that we have verified that the template looks ok, then we can remove the `--first <lines>` flag from the `repeatcsv` command and the dry flag can be removed so that the events are created in Cumulocity IoT.

    <CodeExample>
    
    ```bash
    c8y util repeatcsv machine_events.csv \
    | c8y events create \
        --template "{time: _.Now(input.value.timestamp), text: input.value.text, type: 'machine_CustomEvent'}" \
        --device 12345
    ```

    </CodeExample>

    ```text title="Output"
    | id           | type                     | text                                                            | source.id  | source.name     | time                               | creationTime |
    |--------------|--------------------------|-----------------------------------------------------------------|------------|-----------------|------------------------------------|--------------|
    | 2211591      | machine_CustomEvent      | Alarm Reset drücken AKTIV<br>Alarm Rücktür AAA offen AKTIV      | 12345      | TestDevice      | 2023-06-10T12:00:00.000+02:00      | 2023-06-10T… |
    | 2211592      | machine_CustomEvent      | Alarm Reset drücken INAKTIV<br>Alarm Rücktür AAA offen INAK…    | 12345      | TestDevice      | 2023-06-10T12:10:00.000+02:00      | 2023-06-10T… |
    | 2207833      | machine_CustomEvent      | Alarm Reset drücken GELÖSCHT<br>Alarm Rücktür AAA offen GE…     | 12345      | TestDevice      | 2023-06-10T12:20:00.000+02:00      | 2023-06-10T… |
    | 2211593      | machine_CustomEvent      | Alarm Sattelheizung1: Stoppen aktiv AKTIV                       | 12345      | TestDevice      | 2023-06-10T12:30:00.000+02:00      | 2023-06-10T… |
    ```
