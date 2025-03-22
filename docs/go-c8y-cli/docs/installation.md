---
category: Installation
title: Installation
id: installation
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import DocCardList from '@theme/DocCardList';


### MacOS / Linux / WSL2

<Tabs
    groupId="host_os"
    defaultValue="wget"
    values={[
        { label: 'wget', value: 'wget' },
        { label: 'curl', value: 'curl' },
    ]
    }>
    <TabItem value="wget">

    ```bash
    wget -qO - https://goc8ycli.netlify.app/install.sh | sh -s
    ```

    </TabItem>
    <TabItem value="curl">

    ```bash
    curl -fsSL https://goc8ycli.netlify.app/install.sh | sh -s
    ```

    </TabItem>
</Tabs>


### Windows

```bash
Invoke-Expression (Invoke-WebRequest https://docs-easy-installer--goc8ycli.netlify.app/install.ps1).RawContent
```

### Alternatives

<DocCardList />
