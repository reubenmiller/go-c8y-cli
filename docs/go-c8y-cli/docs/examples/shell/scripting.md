---
title: Bash scripting examples
---

## Reading from file

### Iterate over a file

Read an input file containing a list of device managed object ids (one per line), and fetch updated information about the device (name, type and lastUpdated)

```json title="file: input.list"
11111
22222
33333
```

```bash
# print csv header
echo "id,name,type,lastUpdated"

cat "input.list" | while read line
do
    local current_id="$line"
    c8y devices get -n --id "$current_id" --select id,name,type,lastUpdated --output csv

    local mo=$( c8y devices get -n --id "$current_id" )
    local current_name=$( echo "$mo" | jq -r '.name' )
    local current_type=$( echo "$mo" | jq -r '.type' )
    local last_updated=$( echo "$mo" | jq -r '.lastUpdated' )
    local last_updated_utc=$( date --date "$last_updated" --iso-8601=seconds --universal )

    echo "$current_id,$current_name,$last_updated_utc"
done
```

:::tip
c8y automatically detects and reads from standard input. When using read inside a loop, it will write to standard input, `c8y` will intercept the input and the loop will only be run once instead of once per line.

To get around this, use the `-n/--nullInput` parameter. This parameter will disable reading from standard input.
:::

### Running code on each result

```
i=0
while read line; do
  ((i++))
  name="$( echo "$line" | jq -r '.name' )"
  echo "line $1: $name"
  
done < <( c8y devices list --pageSize 10 | c8y devices get --delay 1000 )
```

### Classifying types of devices into different groups using grep

The following gets a list of files, and filters the json output into different files based on a filter criteria using grep.

:::note
`grep` is being applied to the devices managed object in json format, so the criteria will match any text (not just property values).

If you want to check a specific value then use `jq`
:::

```bash
c8y devices list --pageSize 1000 |
    tee >(grep type1 > match1.out) \
        >(grep type2 > match2.out) \
        >(grep type3 > match3.out) > /dev/null

# Show number of matches per 
echo "match1.out: matches: $(cat match1.out | wc -l)"
echo "match2.out: matches: $(cat match2.out | wc -l)"
echo "match3.out: matches: $(cat match3.out | wc -l)"
```

The same snippet could be also be written to use `jq` instead of `grep`, but this time the criteria is matching by the type property of each device.

```bash
c8y devices list --pageSize 1000 |
    tee \
        >(jq '. | select(.type == "ci_Test")' > match1.out) \
        >(jq '. | select(.type == "type2")' > match2.out) \
        >(jq '. | select(.type == "type3")' > match3.out) > /dev/null

# Show number of matches per 
echo "match1.out: matches: $(cat match1.out | wc -l)"
echo "match2.out: matches: $(cat match2.out | wc -l)"
echo "match3.out: matches: $(cat match3.out | wc -l)"
```

### Backup EPL Monitor files from the Streaming Analytics engine

You can easily backup EPL Monitor files by exporting them from Cumulocity IoT and saving them to disk.

Each EPL Monitor file is saved as a separate file using it's name.

:::Note
This example requires `jq` to be installed.
:::

```bash
#!/bin/bash
#
# Save EPL Monitor files in the platform from the Streaming Analytics engine
#
# Export each monitor file using the name as the the file name (with .mon) extension
#
i=0

while read -r line; do
    ((i++))

    name="$( echo "$line" | jq -r '.name' )"
    OUTPUT_FILE="${name}.mon"

    echo "Saving EPL app ($i): $OUTPUT_FILE"
    echo "$line" | jq -r '.apama_eplfile.contents' > "$OUTPUT_FILE"
  
done < <( c8y inventory list --type "apama_eplfile" --includeAll )
```
