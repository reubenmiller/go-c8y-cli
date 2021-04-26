---
title: Applications - Web (Hosted)
---

import CodeExample from '@site/src/components/CodeExample';

## Hosted (web) Applications


### Create or update a hosted application


<CodeExample>

```bash
# Create from a build folder
c8y applications createHostedApplication --name "myapp" --file "./build"

# From a zip file (name will default to the zip filename)
c8y applications createHostedApplication --file "./build/myapp.zip"
```

</CodeExample>

:::info
The command will check if the application already exists or not, and create the application if necessary. Otherwise it just uploads the binary and activates it.

If you don't want to activate the binary use the `skipActivation` parameter
:::
