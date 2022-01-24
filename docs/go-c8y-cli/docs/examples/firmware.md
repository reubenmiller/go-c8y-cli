---
title: Firmware
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Firmware

### Download the binaries for all versions for a specific version


<CodeExample transform="false">

```bash
while read -r line;
do
    outputFile=$( echo "$line" | cut -d, -f2);
    url=$( echo "$line" | cut -d, -f1);
    c8y api --url "$url" -n --timeout "30m" > "${outputFile}";
done < <(
    c8y firmware versions list \
        --firmware linux-iot \
        --select "c8y_Firmware.url,childAdditions.references.0.managedObject.name" \
        --pageSize 100 \
        -o csv
)
```

</CodeExample>


### Download all of the binaries for all firmware and versions

The above example can be extended to download the binaries of each version of each firmware.

```bash
while read -r line;
do
    outputFile=$( echo "$line" | cut -d, -f2);
    url=$( echo "$line" | cut -d, -f1);
    c8y api --url "$url" -n --timeout "30m" > "${outputFile}";
done < <(
    c8y firmware list -p 100 \
    | c8y firmware versions list \
        --filter "c8y_Firmware.url like https*" \
        --select "c8y_Firmware.url,childAdditions.references.0.managedObject.name" \
        --pageSize 100 \
        -o csv
)
```
